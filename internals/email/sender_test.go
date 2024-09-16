package email

import (
	"email-blaze/internals/config"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSMTPClient is a mock implementation of the SMTP client
type MockSMTPClient struct {
	mock.Mock
}

func (m *MockSMTPClient) Auth(auth interface{}) error {
	args := m.Called(auth)
	return args.Error(0)
}

func (m *MockSMTPClient) Mail(from string, options interface{}) error {
	args := m.Called(from, options)
	return args.Error(0)
}

func (m *MockSMTPClient) Rcpt(to string, options interface{}) error {
	args := m.Called(to, options)
	return args.Error(0)
}

func (m *MockSMTPClient) Data() (interface{}, error) {
	args := m.Called()
	return args.Get(0), args.Error(1)
}

func (m *MockSMTPClient) Quit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSMTPClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestSend(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "test@example.com",
		SMTPPassword: "password123",
	}

	sender := NewSender(cfg)

	mockClient := new(MockSMTPClient)
	mockClient.On("Auth", mock.Anything).Return(nil)
	mockClient.On("Mail", mock.Anything, mock.Anything).Return(nil)
	mockClient.On("Rcpt", mock.Anything, mock.Anything).Return(nil)
	mockClient.On("Data").Return(struct{}{}, nil)
	mockClient.On("Quit").Return(nil)
	mockClient.On("Close").Return(nil)

	// Test successful send
	err := sender.Send("from@example.com", "to@example.com", "Test Subject", "Test Body", false)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestSendWithVerifiedSender(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "test@example.com",
		SMTPPassword: "password123",
	}

	sender := NewSender(cfg)

	mockClient := new(MockSMTPClient)
	mockClient.On("Auth", mock.Anything).Return(nil)
	mockClient.On("Mail", mock.Anything, mock.Anything).Return(nil)
	mockClient.On("Rcpt", mock.Anything, mock.Anything).Return(nil)
	mockClient.On("Data").Return(struct{}{}, nil)
	mockClient.On("Quit").Return(nil)
	mockClient.On("Close").Return(nil)

	// Test successful send with verified sender
	err := sender.SendWithVerifiedSender("from@example.com", "to@example.com", "Test Subject", "Test Body", "reply@example.com")
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestSendRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request SendRequest
		wantErr bool
	}{
		{
			name: "Valid request",
			request: SendRequest{
				From:    "from@example.com",
				To:      "to@example.com",
				Subject: "Test Subject",
				Body:    "Test Body",
			},
			wantErr: false,
		},
		{
			name: "Subject too long",
			request: SendRequest{
				From:    "from@example.com",
				To:      "to@example.com",
				Subject: string(make([]byte, 79)),
				Body:    "Test Body",
			},
			wantErr: true,
		},
		{
			name: "Body too large",
			request: SendRequest{
				From:    "from@example.com",
				To:      "to@example.com",
				Subject: "Test Subject",
				Body:    string(make([]byte, 1000001)),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParse(t *testing.T) {
	emailContent := `From: sender@example.com
To: recipient@example.com
Subject: =?UTF-8?Q?Test_Subject_with_Encoded_Characters_=C3=A9=C3=A0=C3=B9?=
Content-Type: text/plain; charset="UTF-8"

This is the email body.
It can span multiple lines.`

	reader := strings.NewReader(emailContent)
	email, err := Parse(reader)

	assert.NoError(t, err)
	assert.Equal(t, "sender@example.com", email.From)
	assert.Equal(t, []string{"recipient@example.com"}, email.To)
	assert.Equal(t, "Test Subject with Encoded Characters éàù", email.Subject)
	assert.Equal(t, "This is the email body.\nIt can span multiple lines.", email.Body)
}

func TestDecodeRFC2047(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "UTF-8 Encoded",
			input:    "=?UTF-8?Q?Test_Subject_with_Encoded_Characters_=C3=A9=C3=A0=C3=B9?=",
			expected: "Test Subject with Encoded Characters éàù",
			hasError: false,
		},
		{
			name:     "No Encoding",
			input:    "Simple Subject",
			expected: "Simple Subject",
			hasError: false,
		},
		{
			name:     "Invalid Encoding",
			input:    "=?Invalid?Q?Test?=",
			expected: "=?Invalid?Q?Test?=",
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := decodeRFC2047(tc.input)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expected, result)
		})
	}
}
