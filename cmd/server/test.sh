#!/bin/bash

# SMTP server details
SMTP_SERVER="localhost"
SMTP_PORT="25"

# Email details
FROM_EMAIL="support@pictify.io"
TO_EMAIL="suyash@pictify.io"
SUBJECT="Test Email from SMTP Server"
BODY="This is a test email sent using netcat to verify the SMTP server functionality."

# Use netcat to connect to the SMTP server and send the email
{
    sleep 1
    echo "EHLO localhost"
    sleep 1
    echo "MAIL FROM:<$FROM_EMAIL>"
    sleep 1
    echo "RCPT TO:<$TO_EMAIL>"
    sleep 1
    echo "DATA"
    sleep 1
    echo "From: $FROM_EMAIL"
    echo "To: $TO_EMAIL"
    echo "Subject: $SUBJECT"
    echo ""
    echo "$BODY"
    echo "."
    sleep 1
    echo "QUIT"
} | nc $SMTP_SERVER $SMTP_PORT

echo "Email sending attempt completed. Check the SMTP server logs for details."