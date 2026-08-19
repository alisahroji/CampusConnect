package main

import (
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
)

// SendOTPEmail berfungsi untuk mengirimkan kode verifikasi ke email asli user
func SendOTPEmail(toEmail string, otpCode string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	client := resend.NewClient(apiKey)

	// Menyusun isi pesan email
	params := &resend.SendEmailRequest{
		From:    "CampusConnect <onboarding@resend.dev>", // Menggunakan domain uji coba resmi Resend
		To:      []string{toEmail},
		Subject: "Kode Verifikasi OTP - CampusConnect",
		Html: fmt.Sprintf(`
			<div style="font-family: Arial, sans-serif; padding: 20px; color: #333;">
				<h2 style="color: #4F46E5;">Selamat Datang di CampusConnect!</h2>
				<p>Berikut adalah kode verifikasi OTP Anda untuk melanjutkan proses:</p>
				<div style="background: #F3F4F6; padding: 15px; font-size: 24px; font-weight: bold; letter-spacing: 5px; text-align: center; width: 200px; border-radius: 8px; margin: 20px 0;">
					%s
				</div>
				<p>Kode ini berlaku selama 5 menit. Jangan berikan kode ini kepada siapa pun.</p>
				<hr style="border: none; border-top: 1px solid #E5E7EB; margin: 20px 0;" />
				<p style="font-size: 12px; color: #6B7280;">Email ini dikirim secara otomatis oleh sistem CampusConnect.</p>
			</div>
		`, otpCode),
	}

	_, err := client.Emails.Send(params)
	if err != nil {
		return err
	}

	return nil
}