# Mail Module

The Mail module provides a flexible email sending system for the GOE framework. It implements the `contracts.Mailer` interface and supports multiple email providers including SMTP, Resend, and AWS SES.

## Features

- Multiple email provider support (SMTP, Resend, AWS SES)
- Fluent API for composing emails
- HTML and plain text email support
- Email attachments
- Queue-based sending for better performance
- Custom headers support

## Usage

### Initialization

The mail module is automatically initialized by the GOE framework if the mailer feature is enabled:

```
# Enable the mailer in your configuration
MAILER_ENABLED=true

# Configure the default provider
MAILER_PROVIDER=smtp  # Options: smtp, resend, ses

# Set the default sender information
MAILER_FROM_EMAIL=noreply@example.com
MAILER_FROM_NAME=Example App
```

### Provider-specific Configuration

#### SMTP Configuration

```
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=user
SMTP_PASSWORD=password
SMTP_TLS=true
SMTP_LOCAL_NAME=localhost
SMTP_AUTH_METHOD=PLAIN  # Options: PLAIN, LOGIN, CRAM-MD5
```

#### Resend Configuration

```
RESEND_API_KEY=your_api_key
```

#### AWS SES Configuration

```
SES_REGION=us-west-2
SES_ACCESS_KEY_ID=your_access_key
SES_SECRET_ACCESS_KEY=your_secret_key
SES_ENDPOINT=https://email.us-west-2.amazonaws.com  # Optional
```

### Sending Emails

```go
// Get the mailer instance
mailer := goe.UseMailer()

// Send an email using the default provider
err := mailer.DefaultSender().
    To(&[]*mail.Address{
        {Name: "John Doe", Address: "john@example.com"},
    }).
    Subject("Hello from GOE").
    HTML("<h1>Hello World</h1><p>This is a test email from GOE.</p>").
    Send()

// Send an email using a specific provider
err := mailer.GetSender(contracts.ProviderResend).
    To(&[]*mail.Address{
        {Name: "John Doe", Address: "john@example.com"},
    }).
    Cc(&[]*mail.Address{
        {Name: "Jane Doe", Address: "jane@example.com"},
    }).
    Bcc(&[]*mail.Address{
        {Name: "Admin", Address: "admin@example.com"},
    }).
    Subject("Important Notification").
    HTML("<h1>Important Update</h1><p>Please read this important update.</p>").
    Text("Important Update\n\nPlease read this important update.").
    Headers(map[string]string{
        "X-Priority": "1",
    }).
    Attachments(map[string]string{
        "report.pdf": "/path/to/report.pdf",
    }).
    Send(true)  // true to use queue
```

### Adding Custom Providers

You can register custom email providers:

```go
// Create a factory function for your custom provider
factory := func() (contracts.EmailProvider, error) {
    return &MyCustomProvider{}, nil
}

// Register the provider
mailer.RegisterProvider("custom", factory)

// Use the custom provider
mailer.GetSender("custom").
    To(&recipients).
    Subject("Test").
    HTML("<p>Test</p>").
    Send()
```

## Implementation Details

The mail module uses a provider-based architecture that allows for easy switching between different email services. It includes:

- A fluent API for composing emails
- Support for HTML and plain text content
- Automatic HTML to text conversion if only HTML is provided
- Queue-based sending for better performance
- Attachment support
- Custom headers

The module is designed to be extensible, allowing for easy addition of new email providers.