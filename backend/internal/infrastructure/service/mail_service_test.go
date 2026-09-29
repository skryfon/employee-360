package service_test

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/skryfon/employee360/backend/config"
	domainservice "github.com/skryfon/employee360/backend/internal/domain/service"
	infraservice "github.com/skryfon/employee360/backend/internal/infrastructure/service"
)

// mockSMTPServer implements a lightweight in-memory RFC 5321 SMTP server for unit tests.
type mockSMTPServer struct {
	listener     net.Listener
	addr         string
	port         int
	receivedMsgs [][]byte
	mu           sync.Mutex
	requireAuth  bool
	authUsername string
	authPassword string
	closed       bool
}

func startMockSMTPServer(t *testing.T, requireAuth bool, username, password string) *mockSMTPServer {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on test tcp address: %v", err)
	}

	_, portStr, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		t.Fatalf("failed to parse port: %v", err)
	}
	port, _ := strconv.Atoi(portStr)

	server := &mockSMTPServer{
		listener:     l,
		addr:         "127.0.0.1",
		port:         port,
		requireAuth:  requireAuth,
		authUsername: username,
		authPassword: password,
	}

	go server.serve()

	return server
}

func (s *mockSMTPServer) close() {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
	_ = s.listener.Close()
}

func (s *mockSMTPServer) getReceivedMessages() [][]byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := make([][]byte, len(s.receivedMsgs))
	copy(copied, s.receivedMsgs)
	return copied
}

func (s *mockSMTPServer) serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.mu.Lock()
			closed := s.closed
			s.mu.Unlock()
			if closed {
				return
			}
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *mockSMTPServer) handleConn(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	writer := textproto.NewWriter(bufio.NewWriter(conn))

	// Initial greeting
	_ = writer.PrintfLine("220 mock.employee360.local ESMTP MockServer")

	var inData bool
	var dataBuffer strings.Builder
	authenticated := !s.requireAuth

	for {
		if inData {
			line, err := tp.ReadLine()
			if err != nil {
				return
			}
			if line == "." {
				inData = false
				s.mu.Lock()
				s.receivedMsgs = append(s.receivedMsgs, []byte(dataBuffer.String()))
				s.mu.Unlock()
				dataBuffer.Reset()
				_ = writer.PrintfLine("250 2.0.0 OK: queued")
				continue
			}
			// Unescape leading dot
			if strings.HasPrefix(line, "..") {
				line = line[1:]
			}
			dataBuffer.WriteString(line + "\r\n")
			continue
		}

		line, err := tp.ReadLine()
		if err != nil {
			return
		}

		cmd := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(cmd, "EHLO") || strings.HasPrefix(cmd, "HELO"):
			if s.requireAuth {
				_ = writer.PrintfLine("250-mock.employee360.local at your service")
				_ = writer.PrintfLine("250-AUTH PLAIN")
				_ = writer.PrintfLine("250 8BITMIME")
			} else {
				_ = writer.PrintfLine("250-mock.employee360.local at your service")
				_ = writer.PrintfLine("250 8BITMIME")
			}
		case strings.HasPrefix(cmd, "AUTH PLAIN"):
			parts := strings.SplitN(line, " ", 3)
			if len(parts) >= 3 {
				authenticated = true
				_ = writer.PrintfLine("235 2.7.0 Authentication successful")
			} else {
				_ = writer.PrintfLine("334 ")
				_, _ = tp.ReadLine()
				authenticated = true
				_ = writer.PrintfLine("235 2.7.0 Authentication successful")
			}
		case strings.HasPrefix(cmd, "MAIL FROM:"):
			if s.requireAuth && !authenticated {
				_ = writer.PrintfLine("530 5.7.0 Authentication required")
			} else {
				_ = writer.PrintfLine("250 2.1.0 Sender OK")
			}
		case strings.HasPrefix(cmd, "RCPT TO:"):
			_ = writer.PrintfLine("250 2.1.5 Recipient OK")
		case cmd == "DATA":
			inData = true
			_ = writer.PrintfLine("354 Start mail input; end with <CRLF>.<CRLF>")
		case cmd == "QUIT":
			_ = writer.PrintfLine("221 2.0.0 Bye")
			return
		case cmd == "NOOP" || cmd == "RSET":
			_ = writer.PrintfLine("250 2.0.0 OK")
		default:
			_ = writer.PrintfLine("500 5.5.1 Command unrecognized")
		}
	}
}

// helper to parse and extract text & HTML parts from raw multipart email
func parseMIMEParts(t *testing.T, rawEmail []byte) (textBody, htmlBody string) {
	t.Helper()

	msg, err := mail.ReadMessage(bytes.NewReader(rawEmail))
	if err != nil {
		t.Fatalf("failed to parse MIME email: %v", err)
	}

	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		t.Fatalf("failed to parse Content-Type: %v", err)
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		t.Fatalf("expected multipart message, got %s", mediaType)
	}

	mr := multipart.NewReader(msg.Body, params["boundary"])
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("error reading multipart: %v", err)
		}

		partType, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil {
			t.Fatalf("error parsing part content type: %v", err)
		}

		var reader io.Reader = part
		if part.Header.Get("Content-Transfer-Encoding") == "quoted-printable" {
			reader = quotedprintable.NewReader(part)
		}

		content, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("error reading part body: %v", err)
		}

		if partType == "text/plain" {
			textBody = string(content)
		} else if partType == "text/html" {
			htmlBody = string(content)
		}
	}

	return textBody, htmlBody
}

// Ensure SMTPMailService implements domainservice.EmailService at compile time.
var _ domainservice.EmailService = (*infraservice.SMTPMailService)(nil)

func TestSMTPMailService_Send_PasswordReset(t *testing.T) {
	server := startMockSMTPServer(t, false, "", "")
	defer server.close()

	cfg := config.SMTPConfig{
		Host:     server.addr,
		Port:     server.port,
		Username: "",
		Password: "",
		From:     "Employee360 <no-reply@employee360.local>",
		UseTLS:   false,
	}

	mailService := infraservice.NewSMTPMailService(cfg)

	msg := domainservice.EmailMessage{
		To:           "jane.doe@example.com",
		TemplateName: domainservice.EmailTemplatePasswordReset,
		TemplateData: map[string]interface{}{
			"ResetURL":   "https://app.employee360.example/reset-password?token=secret123",
			"TenantName": "Acme Corp",
			"ExpiresIn":  "15 minutes",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mailService.Send(ctx, msg); err != nil {
		t.Fatalf("expected Send() to succeed, got: %v", err)
	}

	msgs := server.getReceivedMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 received email, got %d", len(msgs))
	}

	raw := string(msgs[0])

	// Verify headers
	if !strings.Contains(raw, "To: jane.doe@example.com") {
		t.Errorf("missing To header in received mail: %s", raw)
	}
	if !strings.Contains(raw, "From: Employee360 <no-reply@employee360.local>") {
		t.Errorf("missing From header in received mail: %s", raw)
	}
	if !strings.Contains(raw, "Subject: Reset your Employee360 password") {
		t.Errorf("missing Subject in received mail: %s", raw)
	}
	if !strings.Contains(raw, "multipart/alternative") {
		t.Errorf("expected multipart/alternative Content-Type: %s", raw)
	}

	// Parse and verify decoded parts
	textBody, htmlBody := parseMIMEParts(t, msgs[0])

	if !strings.Contains(textBody, "https://app.employee360.example/reset-password?token=secret123") {
		t.Errorf("expected ResetURL in text body: %s", textBody)
	}
	if !strings.Contains(textBody, "Acme Corp") {
		t.Errorf("expected TenantName in text body: %s", textBody)
	}
	if !strings.Contains(textBody, "15 minutes") {
		t.Errorf("expected ExpiresIn in text body: %s", textBody)
	}

	if !strings.Contains(htmlBody, "https://app.employee360.example/reset-password?token=secret123") {
		t.Errorf("expected ResetURL in html body: %s", htmlBody)
	}
	if !strings.Contains(htmlBody, "Acme Corp") {
		t.Errorf("expected TenantName in html body: %s", htmlBody)
	}
}

func TestSMTPMailService_Send_UserInvitation(t *testing.T) {
	server := startMockSMTPServer(t, false, "", "")
	defer server.close()

	cfg := config.SMTPConfig{
		Host:     server.addr,
		Port:     server.port,
		Username: "",
		Password: "",
		From:     "Employee360 <no-reply@employee360.local>",
		UseTLS:   false,
	}

	mailService := infraservice.NewSMTPMailService(cfg)

	msg := domainservice.EmailMessage{
		To:           "john.smith@example.com",
		TemplateName: domainservice.EmailTemplateUserInvitation,
		TemplateData: map[string]interface{}{
			"invite_url":  "https://admin.employee360.example/invitation?token=invite789",
			"tenant_name": "Globex Corp",
			"role_name":   "HR Admin",
			"expires_at":  "2026-10-05",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mailService.Send(ctx, msg); err != nil {
		t.Fatalf("expected Send() to succeed, got: %v", err)
	}

	msgs := server.getReceivedMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 received email, got %d", len(msgs))
	}

	raw := string(msgs[0])

	if !strings.Contains(raw, "To: john.smith@example.com") {
		t.Errorf("missing To header in received mail: %s", raw)
	}
	if !strings.Contains(raw, "Subject: You have been invited to Employee360") {
		t.Errorf("missing Subject in received mail: %s", raw)
	}

	textBody, htmlBody := parseMIMEParts(t, msgs[0])

	if !strings.Contains(textBody, "https://admin.employee360.example/invitation?token=invite789") {
		t.Errorf("expected InviteURL in text body: %s", textBody)
	}
	if !strings.Contains(textBody, "Globex Corp") {
		t.Errorf("expected TenantName in text body: %s", textBody)
	}
	if !strings.Contains(textBody, "HR Admin") {
		t.Errorf("expected RoleName in text body: %s", textBody)
	}

	if !strings.Contains(htmlBody, "https://admin.employee360.example/invitation?token=invite789") {
		t.Errorf("expected InviteURL in html body: %s", htmlBody)
	}
	if !strings.Contains(htmlBody, "Globex Corp") {
		t.Errorf("expected TenantName in html body: %s", htmlBody)
	}
	if !strings.Contains(htmlBody, "HR Admin") {
		t.Errorf("expected RoleName in html body: %s", htmlBody)
	}
}

func TestSMTPMailService_Send_WithAuth(t *testing.T) {
	server := startMockSMTPServer(t, true, "resend", "re_test_key_123")
	defer server.close()

	cfg := config.SMTPConfig{
		Host:     server.addr,
		Port:     server.port,
		Username: "resend",
		Password: "re_test_key_123",
		From:     "Employee360 <no-reply@example.com>",
		UseTLS:   false,
	}

	mailService := infraservice.NewSMTPMailService(cfg)

	msg := domainservice.EmailMessage{
		To:           "authenticated.user@example.com",
		TemplateName: domainservice.EmailTemplatePasswordReset,
		TemplateData: map[string]interface{}{
			"ResetURL": "https://example.com/reset",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mailService.Send(ctx, msg); err != nil {
		t.Fatalf("expected Send() with auth to succeed, got: %v", err)
	}

	msgs := server.getReceivedMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 received email, got %d", len(msgs))
	}
}

func TestSMTPMailService_Send_CustomSubject(t *testing.T) {
	server := startMockSMTPServer(t, false, "", "")
	defer server.close()

	cfg := config.SMTPConfig{
		Host:   server.addr,
		Port:   server.port,
		From:   "no-reply@example.com",
		UseTLS: false,
	}

	mailService := infraservice.NewSMTPMailService(cfg)

	msg := domainservice.EmailMessage{
		To:           "custom.subject@example.com",
		Subject:      "Important: Your Password Reset Request for {{.TenantName}}",
		TemplateName: domainservice.EmailTemplatePasswordReset,
		TemplateData: map[string]interface{}{
			"ResetURL":   "https://example.com/reset",
			"TenantName": "Stark Industries",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mailService.Send(ctx, msg); err != nil {
		t.Fatalf("expected Send() to succeed, got: %v", err)
	}

	msgs := server.getReceivedMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 received email, got %d", len(msgs))
	}

	raw := string(msgs[0])
	if !strings.Contains(raw, "Subject: Important: Your Password Reset Request for Stark Industries") {
		t.Errorf("expected rendered custom subject, got raw: %s", raw)
	}
}

func TestSMTPMailService_Send_ValidationErrors(t *testing.T) {
	mailService := infraservice.NewSMTPMailService(config.SMTPConfig{
		Host: "localhost",
		Port: 1025,
	})

	ctx := context.Background()

	// Empty To
	err := mailService.Send(ctx, domainservice.EmailMessage{
		To:           "",
		TemplateName: domainservice.EmailTemplatePasswordReset,
	})
	if err == nil || !strings.Contains(err.Error(), "recipient") {
		t.Errorf("expected error for empty recipient, got: %v", err)
	}

	// Empty TemplateName
	err = mailService.Send(ctx, domainservice.EmailMessage{
		To:           "user@example.com",
		TemplateName: "",
	})
	if err == nil || !strings.Contains(err.Error(), "template name") {
		t.Errorf("expected error for empty template name, got: %v", err)
	}

	// Unknown TemplateName
	err = mailService.Send(ctx, domainservice.EmailMessage{
		To:           "user@example.com",
		TemplateName: "NonExistentTemplate",
	})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error for non-existent template, got: %v", err)
	}
}

func TestSMTPMailService_Send_ContextCancellation(t *testing.T) {
	// Create a listener that accepts and hangs without responding to test timeout/cancellation
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer l.Close()

	_, portStr, _ := net.SplitHostPort(l.Addr().String())
	port, _ := strconv.Atoi(portStr)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			// Do nothing with connection to simulate hanging server
			_ = conn
		}
	}()

	cfg := config.SMTPConfig{
		Host:   "127.0.0.1",
		Port:   port,
		From:   "no-reply@example.com",
		UseTLS: false,
	}

	mailService := infraservice.NewSMTPMailService(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = mailService.Send(ctx, domainservice.EmailMessage{
		To:           "user@example.com",
		TemplateName: domainservice.EmailTemplatePasswordReset,
		TemplateData: map[string]interface{}{
			"ResetURL": "https://example.com/reset",
		},
	})

	if err == nil {
		t.Fatal("expected error on context cancellation/timeout, got nil")
	}
}
