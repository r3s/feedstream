package service

import (
	"fmt"
	"log"
	"rss-reader/internal/domain"
	"rss-reader/internal/repository"
	"rss-reader/pkg/email"
	"rss-reader/pkg/ratelimit"
	"rss-reader/pkg/security"
	"time"
)

type AuthService struct {
	userRepo      repository.UserRepository
	otpRepo       repository.OTPRepository
	emailService  email.Service
	otpGenerator  *security.OTPGenerator
	sendLimiter   *ratelimit.Limiter
	verifyLimiter *ratelimit.Limiter
}

func NewAuthService(
	userRepo repository.UserRepository,
	otpRepo repository.OTPRepository,
	emailService email.Service,
	otpGenerator *security.OTPGenerator,
) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		otpRepo:       otpRepo,
		emailService:  emailService,
		otpGenerator:  otpGenerator,
		sendLimiter:   ratelimit.NewLimiter(),
		verifyLimiter: ratelimit.NewLimiter(),
	}
}

func (s *AuthService) SendOTP(email string) error {
	if !s.sendLimiter.Allow(email, 3, 15*time.Minute) {
		log.Printf("Rate limit exceeded for OTP send to: %s", email)
		return fmt.Errorf("too many OTP requests, please try again in 15 minutes")
	}

	_, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			_, err = s.userRepo.Create(email)
			if err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
			log.Printf("Created new user with email: %s", email)
		} else {
			return fmt.Errorf("failed to get user: %w", err)
		}
	}

	otp, err := s.otpGenerator.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	if err := s.otpRepo.DeleteByEmail(email); err != nil {
		log.Printf("Warning: failed to delete old OTPs for %s: %v", email, err)
	}

	expiresAt := time.Now().Add(10 * time.Minute)
	if err := s.otpRepo.Store(email, otp, expiresAt); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	subject := "Your OTP for FeedStream RSS Reader"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; padding: 20px;">
			<p>Your OTP code is:</p>
			<p style="font-size: 32px; font-weight: bold; color: #333; letter-spacing: 4px; margin: 20px 0;">%s</p>
			<p style="color: #666;">This code will expire in 10 minutes.</p>
		</body>
		</html>
	`, otp)
	if err := s.emailService.SendEmail(email, subject, body); err != nil {
		log.Printf("Error sending OTP email to %s: %v", email, err)
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	log.Printf("OTP sent successfully to: %s", email)
	return nil
}

func (s *AuthService) VerifyOTP(email, otpCode string) (*domain.User, error) {
	if !s.verifyLimiter.Allow(email, 5, 15*time.Minute) {
		log.Printf("Rate limit exceeded for OTP verification: %s", email)
		return nil, fmt.Errorf("too many verification attempts, please wait 15 minutes and request a new OTP")
	}

	storedOTP, err := s.otpRepo.GetLatestByEmail(email)
	if err != nil {
		if err == domain.ErrOTPNotFound {
			return nil, domain.ErrInvalidOTP
		}
		return nil, fmt.Errorf("failed to get OTP: %w", err)
	}

	if !storedOTP.IsValid(otpCode) {
		if storedOTP.IsExpired() {
			return nil, domain.ErrOTPExpired
		}
		return nil, domain.ErrInvalidOTP
	}

	if err := s.otpRepo.DeleteByEmail(email); err != nil {
		log.Printf("Warning: failed to delete OTP for %s: %v", email, err)
	}

	s.verifyLimiter.Reset(email)

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	log.Printf("User %s authenticated successfully", email)
	return user, nil
}

func (s *AuthService) GetUserByID(userID int) (*domain.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}
