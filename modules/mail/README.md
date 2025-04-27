# Mail Module

The Mail module provides a flexible email sending system for the GOE framework. It implements the [`contracts.Mailer`](https://github.com/oeasenet/goe/blob/main/contracts/mail.go) interface and supports multiple email providers including SMTP, Resend, and AWS SES.

## Features

- Multiple email provider support (SMTP, Resend, AWS SES)
- Fluent API for composing emails
- HTML and plain text email support
- Automatic HTML to text conversion
- Email attachments with file path support
- Queue-based sending for better performance and reliability
- Custom headers support
- Thread-safe operations
- Extensible provider architecture
- Comprehensive error handling

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

#### Basic Email

```go
// Get the mailer instance
mailer := goe.UseMailer()

// Send a simple email
err := mailer.DefaultSender().
    To(&[]*mail.Address{
        {Name: "John Doe", Address: "john@example.com"},
    }).
    Subject("Hello from GOE").
    HTML("<h1>Hello World</h1><p>This is a test email from GOE.</p>").
    Send()

if err != nil {
    // Handle error
    goe.UseLog().Error("Failed to send email:", err)
}
```

#### Advanced Email with Attachments and CC/BCC

```go
// Get the mailer instance
mailer := goe.UseMailer()

// Create recipients
recipients := &[]*mail.Address{
    {Name: "John Doe", Address: "john@example.com"},
    {Name: "Jane Smith", Address: "jane@example.com"},
}

ccRecipients := &[]*mail.Address{
    {Name: "Manager", Address: "manager@example.com"},
}

bccRecipients := &[]*mail.Address{
    {Name: "Admin", Address: "admin@example.com"},
}

// Send an email with attachments and custom headers
err := mailer.DefaultSender().
    To(recipients).
    Cc(ccRecipients).
    Bcc(bccRecipients).
    Subject("Monthly Report").
    HTML("<h1>Monthly Report</h1><p>Please find attached the monthly report.</p>").
    Text("Monthly Report\n\nPlease find attached the monthly report.").
    Headers(map[string]string{
        "X-Priority": "1",
        "X-Custom-Header": "Custom Value",
    }).
    Attachments(map[string]string{
        "report.pdf": "/path/to/report.pdf",
        "data.xlsx": "/path/to/data.xlsx",
    }).
    Send()

if err != nil {
    // Handle error
    goe.UseLog().Error("Failed to send email with attachments:", err)
}
```

#### Using a Specific Provider

```go
// Get the mailer instance
mailer := goe.UseMailer()

// Send an email using the Resend provider
err := mailer.GetSender(contracts.ProviderResend).
    To(&[]*mail.Address{
        {Name: "John Doe", Address: "john@example.com"},
    }).
    Subject("Important Notification").
    HTML("<h1>Important Update</h1><p>Please read this important update.</p>").
    Send()

if err != nil {
    // Handle error
    goe.UseLog().Error("Failed to send email via Resend:", err)
}
```

#### Queue-based Sending

```go
// Get the mailer instance
mailer := goe.UseMailer()

// Send an email using the queue (for better performance)
err := mailer.DefaultSender().
    To(&[]*mail.Address{
        {Name: "John Doe", Address: "john@example.com"},
    }).
    Subject("Welcome to our platform").
    HTML("<h1>Welcome!</h1><p>Thank you for signing up.</p>").
    Send(true)  // true to use queue

if err != nil {
    // Handle error
    goe.UseLog().Error("Failed to queue email:", err)
}
```

### Real-world Examples

#### User Registration Email

```go
func sendWelcomeEmail(user *User) error {
    mailer := goe.UseMailer()
    
    // Create HTML content with personalization
    htmlContent := fmt.Sprintf(`
        <h1>Welcome to %s, %s!</h1>
        <p>Thank you for creating an account. We're excited to have you on board!</p>
        <p>Your account has been successfully created and is ready to use.</p>
        <p>To get started, please <a href="%s/login">log in</a> to your account.</p>
        <p>If you have any questions, please don't hesitate to contact our support team.</p>
        <p>Best regards,<br>The %s Team</p>
    `, appName, user.Name, appURL, appName)
    
    // Send the welcome email
    return mailer.DefaultSender().
        To(&[]*mail.Address{
            {Name: user.Name, Address: user.Email},
        }).
        Subject(fmt.Sprintf("Welcome to %s!", appName)).
        HTML(htmlContent).
        Send(true)  // Use queue for better performance
}
```

#### Password Reset Email

```go
func sendPasswordResetEmail(user *User, resetToken string) error {
    mailer := goe.UseMailer()
    
    // Create the reset URL
    resetURL := fmt.Sprintf("%s/reset-password?token=%s", appURL, resetToken)
    
    // Create HTML content
    htmlContent := fmt.Sprintf(`
        <h1>Password Reset Request</h1>
        <p>Hello %s,</p>
        <p>We received a request to reset your password. If you didn't make this request, you can safely ignore this email.</p>
        <p>To reset your password, click the link below:</p>
        <p><a href="%s">Reset Your Password</a></p>
        <p>This link will expire in 1 hour.</p>
        <p>Best regards,<br>The %s Team</p>
    `, user.Name, resetURL, appName)
    
    // Send the password reset email
    return mailer.DefaultSender().
        To(&[]*mail.Address{
            {Name: user.Name, Address: user.Email},
        }).
        Subject("Password Reset Request").
        HTML(htmlContent).
        Headers(map[string]string{
            "X-Priority": "1",  // High priority
        }).
        Send(true)  // Use queue for better performance
}
```

### Adding Custom Providers

You can register custom email providers to extend the functionality:

```go
// Define a custom email provider
type MyCustomProvider struct {
    apiKey string
    logger contracts.Logger
}

// Implement the EmailProvider interface
func (p *MyCustomProvider) Send(message *contracts.EmailMessage) error {
    // Implementation for sending emails via your custom provider
    // ...
    return nil
}

func (p *MyCustomProvider) Name() contracts.MailProvider {
    return "custom"
}

// Create a factory function for your custom provider
factory := func() (contracts.EmailProvider, error) {
    return &MyCustomProvider{
        apiKey: "your-api-key",
        logger: goe.UseLog(),
    }, nil
}

// Register the provider
mailer := goe.UseMailer()
err := mailer.RegisterProvider("custom", factory)
if err != nil {
    // Handle error
    goe.UseLog().Error("Failed to register custom provider:", err)
}

// Use the custom provider
err = mailer.GetSender("custom").
    To(&[]*mail.Address{
        {Name: "John Doe", Address: "john@example.com"},
    }).
    Subject("Test").
    HTML("<p>Test</p>").
    Send()
```

## API Reference

### Mailer Interface

The Mail module implements the [`contracts.Mailer`](https://github.com/oeasenet/goe/blob/main/contracts/mail.go) interface:

```go
// Mailer is the main interface for sending emails
type Mailer interface {
    // GetSender returns an EmailSender for the specified provider
    GetSender(provider MailProvider) EmailSender
    
    // DefaultSender returns the default EmailSender
    DefaultSender() EmailSender
    
    // RegisterProvider registers a new email provider
    RegisterProvider(provider MailProvider, factory EmailProviderFactory) error
}
```

### EmailSender Interface

The [`EmailSender`](https://github.com/oeasenet/goe/blob/main/contracts/mail.go#L29) interface provides a fluent API for composing and sending emails:

```go
// EmailSender is the interface for sending emails with a fluent API
type EmailSender interface {
    // To sets the recipients
    To(t *[]*mail.Address) EmailSender
    
    // Bcc sets the BCC recipients
    Bcc(b *[]*mail.Address) EmailSender
    
    // Cc sets the CC recipients
    Cc(c *[]*mail.Address) EmailSender
    
    // Subject sets the subject
    Subject(sub string) EmailSender
    
    // HTML sets the HTML body
    HTML(html string) EmailSender
    
    // Text sets the text body
    Text(text string) EmailSender
    
    // Headers sets the headers
    Headers(h map[string]string) EmailSender
    
    // Attachments sets the attachments
    Attachments(a map[string]string) EmailSender
    
    // Send sends the email
    Send(useQueue ...bool) error
}
```

### EmailProvider Interface

The [`EmailProvider`](https://github.com/oeasenet/goe/blob/main/contracts/mail.go#L55) interface must be implemented by all email providers:

```go
// EmailProvider is the interface that all email providers must implement
type EmailProvider interface {
    // Send sends an email message
    Send(message *EmailMessage) error
    
    // Name returns the name of the provider
    Name() MailProvider
}
```

## Implementation Details

The mail module is implemented in the [`MailerManager`](https://github.com/oeasenet/goe/blob/main/modules/mail/mailer.go) struct, which provides:

- A registry of email providers
- A default provider selection
- Thread-safe operations with mutex protection
- Queue-based sending for better performance

### Provider Architecture

The module uses a provider-based architecture that allows for easy switching between different email services:

1. **SMTP Provider**: For sending emails via SMTP servers
2. **Resend Provider**: For sending emails via the Resend API
3. **SES Provider**: For sending emails via Amazon SES

Each provider implements the `EmailProvider` interface, making it easy to add new providers.

### Queue-based Sending

The module supports queue-based sending for better performance and reliability:

1. When an email is sent with the queue option, it's serialized and pushed to a queue
2. A worker processes the queue and sends the emails asynchronously
3. This prevents email sending from blocking the main application flow
4. It also provides retry capabilities for failed sends

### Attachment Handling

The module handles attachments by:

1. Accepting file paths in the `Attachments` method
2. Converting file paths to `io.Reader` instances when sending
3. For queued emails, storing the file paths and reopening the files when processing the queue

### HTML to Text Conversion

If only HTML content is provided, the module can automatically generate plain text content using the [`html2text`](https://github.com/oeasenet/goe/blob/main/modules/mail/html2text/html2text.go) package.

### Error Handling

The module includes comprehensive error handling:

- Provider initialization errors
- Email sending errors
- Attachment processing errors
- Queue processing errors

All errors are properly logged and can be handled by the application.