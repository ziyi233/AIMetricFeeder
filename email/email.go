package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"html/template"
) // 增加 yaml 包、ioutil、html/template

// Config holds SMTP server configuration
// 推荐通过环境变量配置以下参数：
// EMAIL_SMTP_HOST
// EMAIL_SMTP_PORT
// EMAIL_SMTP_USERNAME
// EMAIL_SMTP_PASSWORD
// EMAIL_SMTP_FROM
// EMAIL_SMTP_USETLS (true/false)
type Config struct {
	Host     string // SMTP server host, e.g. smtp.qq.com
	Port     int    // SMTP server port, e.g. 465 or 587
	Username string // SMTP username (usually your email address)
	Password string // SMTP password or authorization code
	From     string // Sender email address
	UseTLS   bool   // Whether to use TLS (true for 465 or 587)
}

// LoadConfigFromEnv loads email config from environment variables
func LoadConfigFromEnv() (Config, error) {
	port, err := strconv.Atoi(os.Getenv("EMAIL_SMTP_PORT"))
	if err != nil {
		return Config{}, fmt.Errorf("EMAIL_SMTP_PORT 格式错误: %v", err)
	}
	useTLS := os.Getenv("EMAIL_SMTP_USETLS") == "true"
	return Config{
		Host:     os.Getenv("EMAIL_SMTP_HOST"),
		Port:     port,
		Username: os.Getenv("EMAIL_SMTP_USERNAME"),
		Password: os.Getenv("EMAIL_SMTP_PASSWORD"),
		From:     os.Getenv("EMAIL_SMTP_FROM"),
		UseTLS:   useTLS,
	}, nil
}

// LoadConfigFromFileOrEnv 优先从 config.yaml 读取 email 配置，不存在时兜底用环境变量
func LoadConfigFromFileOrEnv() (Config, error) {
	// 1. 先尝试读取 config.yaml
	data, err := ioutil.ReadFile("config.yaml")
	if err == nil {
		var cfgFile struct {
			Email struct {
				Host     string `yaml:"host"`
				Port     int    `yaml:"port"`
				Username string `yaml:"username"`
				Password string `yaml:"password"`
				From     string `yaml:"from"`
				UseTLS   bool   `yaml:"usetls"`
			} `yaml:"email"`
		}
		err = yaml.Unmarshal(data, &cfgFile)
		if err == nil && cfgFile.Email.Host != "" {
			return Config{
				Host:     cfgFile.Email.Host,
				Port:     cfgFile.Email.Port,
				Username: cfgFile.Email.Username,
				Password: cfgFile.Email.Password,
				From:     cfgFile.Email.From,
				UseTLS:   cfgFile.Email.UseTLS,
			}, nil
		}
	}
	// 2. 兜底用环境变量
	return LoadConfigFromEnv()
}

// Email holds the email content
type Email struct {
	To           []string               // Recipient addresses
	Cc           []string               // CC addresses (optional)
	Bcc          []string               // BCC addresses (optional)
	Subject      string                 // Email subject
	Body         string                 // Email body (for plain text)
	IsHTML       bool                   // Whether to use HTML template
	TemplateName string                 // Name of HTML template file (e.g. "default.html")
	TemplateData map[string]interface{} // Data for template rendering
}

// Mailer provides the Send method
 type Mailer struct {
	Config Config
}

// NewMailer creates a new Mailer
func NewMailer(cfg Config) *Mailer {
	return &Mailer{Config: cfg}
}

// Send sends the email
func (m *Mailer) Send(mail Email) error {
	headers := make(map[string]string)
	headers["From"] = m.Config.From
	headers["To"] = strings.Join(mail.To, ", ")
	if len(mail.Cc) > 0 {
		headers["Cc"] = strings.Join(mail.Cc, ", ")
	}
	headers["Subject"] = mail.Subject

	var body string
	if mail.IsHTML {
		headers["Content-Type"] = "text/html; charset=UTF-8"
		// 渲染HTML模板
		if mail.TemplateName != "" {
			tplPath := "email/templates/" + mail.TemplateName
			tplData := mail.TemplateData
			if tplData == nil {
				tplData = make(map[string]interface{})
			}
			tplData["Subject"] = mail.Subject // 默认注入Subject
			content, err := renderHTMLTemplate(tplPath, tplData)
			if err != nil {
				return fmt.Errorf("模板渲染失败: %w", err)
			}
			body = content
		} else {
			body = mail.Body // 兼容直接传HTML
		}
	} else {
		headers["Content-Type"] = "text/plain; charset=UTF-8"
		body = mail.Body
	}

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n" + body)

	recipients := append(mail.To, mail.Cc...)
	recipients = append(recipients, mail.Bcc...)

	addr := fmt.Sprintf("%s:%d", m.Config.Host, m.Config.Port)

	if m.Config.UseTLS {
		tlsconfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         m.Config.Host,
		}
		conn, err := tls.Dial("tcp", addr, tlsconfig)
		if err != nil {
			return err
		}
		c, err := smtp.NewClient(conn, m.Config.Host)
		if err != nil {
			return err
		}
		defer c.Quit()
		if err = c.Auth(smtp.PlainAuth("", m.Config.Username, m.Config.Password, m.Config.Host)); err != nil {
			return err
		}
		if err = c.Mail(m.Config.From); err != nil {
			return err
		}
		for _, addr := range recipients {
			if err = c.Rcpt(addr); err != nil {
				return err
			}
		}
		w, err := c.Data()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(msg.String()))
		if err != nil {
			return err
		}
		return w.Close()
	} else {
		auth := smtp.PlainAuth("", m.Config.Username, m.Config.Password, m.Config.Host)
		return smtp.SendMail(addr, auth, m.Config.From, recipients, []byte(msg.String()))
	}
}

// 渲染HTML模板
func renderHTMLTemplate(tplPath string, data map[string]interface{}) (string, error) {
	tplBytes, err := ioutil.ReadFile(tplPath)
	if err != nil {
		return "", err
	}
	tpl, err := template.New("mail").Parse(string(tplBytes))
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	err = tpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Example usage:
// 推荐使用环境变量方式配置：
// Windows 命令行示例：
// set EMAIL_SMTP_HOST=smtp.qq.com
// set EMAIL_SMTP_PORT=465
// set EMAIL_SMTP_USERNAME=1790060255@qq.com
// set EMAIL_SMTP_PASSWORD=glhtpytpunwtefae
// set EMAIL_SMTP_FROM=1790060255@qq.com
// set EMAIL_SMTP_USETLS=true
//
// func main() {
// 	cfg, err := LoadConfigFromEnv()
// 	if err != nil {
// 		panic(err)
// 	}
// 	mailer := NewMailer(cfg)
// 	email := Email{
// 		To:      []string{"收件人邮箱@example.com"},
// 		Subject: "测试邮件",
// 		Body:    "<h1>你好，这是来自Go的测试邮件！</h1>",
// 		IsHTML:  true,
// 	}
// 	err = mailer.Send(email)
// 	if err != nil {
// 		fmt.Println("发送失败:", err)
// 	} else {
// 		fmt.Println("发送成功！")
// 	}
// }
