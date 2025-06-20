package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/config"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/constants"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/entities"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/censorship"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/customerror"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/encryption"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/random"
	sqltypes "gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/sql_types"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/models"
	emailModule "gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/modules/email"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/repositories"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/values"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// Verify member error
var (
	// Verify member errors
	errMemberWasMissing  = errors.New(customerror.ErrGroupClientErr, "2", "Member pernah hilang")
	errMemberWasUpgraded = errors.New(customerror.ErrGroupClientErr, "3", "Member pernah upgrade")

	// Login errors
	errLoginInvalidCreds  = errors.New(customerror.ErrGroupClientErr, "1", "Kredensial tidak valid")
	errLoginForbidden     = errors.New(customerror.ErrGroupForbiddenErr, "2", "Akses tertutup")
	errInvalidPhoneNumber = errors.New(customerror.ErrGroupClientErr, "1", "Format nomor telepon tidak valid")
	errLoginCooldown      = func(attempts, d time.Duration) errors.Error {
		cooldown := d / time.Minute
		return errors.New(customerror.ErrGroupClientErr, "3", fmt.Sprintf("Anda Telah Gagal Masuk %dx, Mohon Tunggu %d Menit", attempts, cooldown))
	}
	errMemberBlockedOnTooManyLoginAttempt = errors.New(customerror.ErrGroupClientErr, "3", "Akun Anda Telah Ditangguhkan, Silahkan Hubungi Admin")

	// OTP errors
	errInvalidOtp = errors.New(customerror.ErrGroupClientErr, "", "Kode OTP yang anda masukan salah silahkan ulang kembali")
)

// Member mutation status
const (
	memberMutationStatusWasMissing  = 1
	memberMutationStatusWasUpgraded = 2
)

// Consecutive Login attempts
const (
	consecLoginAttempt1 = 5
	consecLoginAttempt2 = 10
	consecLoginAttempt3 = 15
)

const (
	timeUntilLoginIsUnlocked  = "unlocked"
	timeUntilLoginIsPermanent = "permanent"
)

// JWT scope
const (
	JWTScopeFeatures     = "features"
	JWTScopeRefreshToken = "refresh_token"
	JWTScopeEmailSignin  = "email_signin"
)

type AuthService interface {
	VerifyMember(ctx context.Context, memberCode string) (models.VerifyMemberResponse, error)
	Login(ctx context.Context, creds models.LoginRequest) (models.LoginResponse, error)
	GetTimeUntilLoginUnlock(ctx context.Context, memberCode string) (models.GetTimeUntilLoginUnlockResponse, error)
	TokenFromRefreshToken(ctx context.Context, req models.TokenFromRefreshTokenRequest) (models.AuthToken, error)
	RegisterEmailByOTP(ctx context.Context, memberId, memberRegBranchId int, email string) (models.OTPDetail, error)
	ValidateOTPForEmailAndPasswordRegistration(ctx context.Context, memberId, memberRegBranchId int, req models.ValidateOTPForEmailAndPasswordRegistrationRequest) error
	LoginWithEmailAndPassword(ctx context.Context, memberId, memberRegBranchId int, req models.LoginWithEmailAndPasswordRequest) (models.LoginResponse, error)
	RegisterEmailFromGoogleSignin(ctx context.Context, memberId, memberRegBranchId int, idTokenStr string) (models.AuthToken, error)
	LoginWithGoogleSignin(ctx context.Context, memberId, memberRegBranchId int, idTokenStr string) (models.LoginResponse, error)
}

type authService struct {
	authCfg    config.Auth
	memberRepo repositories.MemberRepository
	regOtpRepo repositories.RegistrationOTPRepository
	emailSvc   emailModule.EmailServiceProvider
}

func NewAuthService(authCfg config.Auth, memberRepo repositories.MemberRepository, regOtpRepo repositories.RegistrationOTPRepository, emailSvc emailModule.EmailServiceProvider) *authService {
	return &authService{
		authCfg:    authCfg,
		memberRepo: memberRepo,
		regOtpRepo: regOtpRepo,
		emailSvc:   emailSvc,
	}
}

// VerifyMember is more like "Get member", but with member code/member card number.
func (svc *authService) VerifyMember(ctx context.Context, memberCode string) (models.VerifyMemberResponse, error) {
	membership, err := svc.memberRepo.GetMembershipStatus(ctx, memberCode)
	if err != nil {
		if err == customerror.ErrNotFound {
			return models.VerifyMemberResponse{}, customerror.ErrMemberNotFound
		}

		return models.VerifyMemberResponse{}, err
	}

	switch membership.Status {
	case memberMutationStatusWasMissing:
		return models.VerifyMemberResponse{}, errMemberWasMissing
	case memberMutationStatusWasUpgraded:
		return models.VerifyMemberResponse{}, errMemberWasUpgraded
	}

	memberPhoneNum, err := values.ParseDirtyPhoneNumber(membership.MemberPhoneNumber)
	if err != nil {
		return models.VerifyMemberResponse{}, customerror.ErrInternal.WithError(err).WithSource()
	}

	resp := models.VerifyMemberResponse{
		MemberCode:  memberCode,
		Name:        censorship.PersonNamePartialCensor(membership.MemberName),
		PhoneNumber: censorship.PhoneNumPartialCensor(memberPhoneNum.Standard()),
	}

	return resp, nil
}

func (svc *authService) Login(ctx context.Context, creds models.LoginRequest) (models.LoginResponse, error) {
	membership, err := svc.memberRepo.GetMembershipStatus(ctx, creds.MemberCode)
	if err != nil {
		if err == customerror.ErrNotFound {
			return models.LoginResponse{}, errLoginInvalidCreds
		}

		return models.LoginResponse{}, err
	}

	switch membership.Status {
	case memberMutationStatusWasMissing:
		return models.LoginResponse{}, errMemberWasMissing
	case memberMutationStatusWasUpgraded:
		return models.LoginResponse{}, errMemberWasUpgraded
	}

	loginAttempts, loginAttempted, err := svc.memberRepo.GetFailedLoginAttempt(ctx, membership.MemberID, membership.MemberRegistrationBranchID)
	if err != nil {
		return models.LoginResponse{}, err
	}

	if loginAttempted {
		counter := loginAttempts.Counter
		lastAttempt := loginAttempts.LastAttempt.Time()
		now := time.Now()

		if counter == consecLoginAttempt1 {
			cooldown := time.Duration(svc.authCfg.WaitTimeMinutesOnFailedLogin5Times)
			afterLastAttempt := lastAttempt.Add(cooldown)
			if now.Before(afterLastAttempt) {
				return models.LoginResponse{}, errLoginCooldown(consecLoginAttempt1, time.Duration(svc.authCfg.WaitTimeMinutesOnFailedLogin5Times))
			}
		}

		if counter == consecLoginAttempt2 {
			cooldown := time.Duration(svc.authCfg.WaitTimeMinutesOnFailedLogin15Times)
			afterLastAttempt := lastAttempt.Add(cooldown)
			if now.Before(afterLastAttempt) {
				return models.LoginResponse{}, errLoginCooldown(consecLoginAttempt2, time.Duration(svc.authCfg.WaitTimeMinutesOnFailedLogin15Times))
			}
		}

		if counter >= consecLoginAttempt3 {
			return models.LoginResponse{}, errMemberBlockedOnTooManyLoginAttempt
		}
	}

	inputPhoneNumber, err := values.ParsePhoneNumber(creds.PhoneNumber)
	if err != nil {
		return models.LoginResponse{}, errInvalidPhoneNumber
	}

	memberPhoneNumber, err := values.ParseDirtyPhoneNumber(membership.MemberPhoneNumber)
	if err != nil {
		return models.LoginResponse{}, errLoginForbidden
	}

	if !memberPhoneNumber.IsEqual(inputPhoneNumber) {
		err := svc.memberRepo.IncrFailedLoginAttempt(ctx, membership.MemberID, membership.MemberRegistrationBranchID)
		if err != nil {
			return models.LoginResponse{}, err
		}

		var counter int

		// Check if the counter is still 0
		loginAttempts, _, err := svc.memberRepo.GetFailedLoginAttempt(ctx, membership.MemberID, membership.MemberRegistrationBranchID)
		if err != nil {
			return models.LoginResponse{}, err
		}

		counter = loginAttempts.Counter

		if loginAttempts.Counter == 0 {
			// If counter is indeed still 0, call IncrFailedLoginAttempt once more
			err := svc.memberRepo.IncrFailedLoginAttempt(ctx, membership.MemberID, membership.MemberRegistrationBranchID)
			if err != nil {
				return models.LoginResponse{}, err
			}

			counter++
		}

		// Improve UX by immediately check if the counter already reach a consecutive
		// failed login attempt limit
		switch counter {
		case consecLoginAttempt1:
			return models.LoginResponse{}, errLoginCooldown(consecLoginAttempt1, time.Duration(svc.authCfg.WaitTimeMinutesOnFailedLogin5Times))

		case consecLoginAttempt2:
			return models.LoginResponse{}, errLoginCooldown(consecLoginAttempt2, time.Duration(svc.authCfg.WaitTimeMinutesOnFailedLogin15Times))

		case consecLoginAttempt3:
			return models.LoginResponse{}, errMemberBlockedOnTooManyLoginAttempt
		}

		return models.LoginResponse{}, errLoginInvalidCreds
	}

	if loginAttempted {
		err := svc.memberRepo.ResetLoginAttemptCounter(ctx, membership.MemberID, membership.MemberRegistrationBranchID)
		if err != nil {
			return models.LoginResponse{}, err
		}
	}

	// Check if user has verified email
	emailVerified, err := svc.memberRepo.IsMemberEmailVerified(ctx, membership.MemberID, membership.MemberRegistrationBranchID)
	if err != nil {
		return models.LoginResponse{}, err
	}

	if emailVerified {
		// TODO: Store this value in config
		expiresAtFromNow := 5 * time.Minute
		sub := svc.memberIdAndRegistBranchIdAsJwtSub(membership.MemberID, membership.MemberRegistrationBranchID)
		tokenStr, expiresAt, err := svc.createJwt(JWTScopeEmailSignin, sub, expiresAtFromNow)
		if err != nil {
			return models.LoginResponse{}, err
		}

		return models.LoginResponse{
			EmailVerified: emailVerified,
			AuthToken: models.AuthToken{
				Token:          tokenStr,
				TokenExpiredAt: expiresAt,
			},
		}, err
	}

	authToken, err := svc.createAuthTokenForMember(membership.MemberID, membership.MemberRegistrationBranchID, creds.MemberCode)
	if err != nil {
		return models.LoginResponse{}, err
	}

	return models.LoginResponse{
		EmailVerified: emailVerified,
		AuthToken:     authToken,
	}, nil
}

func (svc *authService) createAuthTokenForMember(memberId, registrationBranchId int, cardNum string) (models.AuthToken, error) {
	var authToken models.AuthToken
	// Previously, we use card number for subject, but it seems like
	// that's a bad idea, since JWT claims should not contains sensitive
	// information, so instead we will use combination of member ID and
	// registration branch ID, concatenate with '-'
	subject := fmt.Sprintf("%d-%d", memberId, registrationBranchId)

	// Create the access token
	jwtStr, tokenExpiredAt, err := svc.createJwt(JWTScopeFeatures, subject, time.Duration(svc.authCfg.JWT.Expiration))
	if err != nil {
		return models.AuthToken{}, err
	}

	authToken.Token = jwtStr
	authToken.TokenExpiredAt = tokenExpiredAt

	// Using card number as subject for the refresh token, because we definitely don't want users
	// with inactive or invalid membership to be able to sign-in to the system. Of course we encrypt
	// it to make sure the card number is safe to transport
	cardNumCipher, err := encryption.AES256Encrypt([]byte(cardNum), svc.authCfg.CardNumberEncryptionKey.Decoded)
	if err != nil {
		return models.AuthToken{}, customerror.ErrInternal.WithError(err).WithSource()
	}

	cardNumCipherBase64 := base64.RawStdEncoding.EncodeToString(cardNumCipher)

	// Create the refresh token.
	refreshJwtStr, refreshTokenExpiredAt, err := svc.createJwt(JWTScopeRefreshToken, cardNumCipherBase64, time.Duration(svc.authCfg.JWT.RefreshTokenExpiration))
	if err != nil {
		return models.AuthToken{}, err
	}

	authToken.RefreshToken = refreshJwtStr
	authToken.RefreshTokenExpiredAt = refreshTokenExpiredAt

	return authToken, nil
}

func (svc *authService) GetTimeUntilLoginUnlock(ctx context.Context, memberCode string) (models.GetTimeUntilLoginUnlockResponse, error) {
	compositeId, err := svc.memberRepo.GetMemberCompositeIDByCardNumber(ctx, memberCode)
	if err != nil {
		return models.GetTimeUntilLoginUnlockResponse{}, err
	}

	loginAttempts, loginAttempted, err := svc.memberRepo.GetFailedLoginAttempt(ctx, compositeId.ID, compositeId.RegistrationBranchID)
	if err != nil {
		if err == customerror.ErrNotFound {
			return models.GetTimeUntilLoginUnlockResponse{Until: timeUntilLoginIsUnlocked}, nil
		}

		return models.GetTimeUntilLoginUnlockResponse{}, err
	}

	var timeUntil time.Time

	if loginAttempted {
		// We increment the counter by 1 because somehow the counter
		// starts from 0, not 1
		counter := loginAttempts.Counter
		lastAttempt := loginAttempts.LastAttempt.Time()
		now := time.Now()

		if counter == consecLoginAttempt1 {
			cooldown := time.Duration(svc.authCfg.WaitTimeMinutesOnFailedLogin5Times)
			afterLastAttempt := lastAttempt.Add(cooldown)
			if now.Before(afterLastAttempt) {
				duration := afterLastAttempt.Sub(now)
				timeUntil = now.Add(duration)
			}
		}

		if counter == consecLoginAttempt2 {
			cooldown := time.Duration(svc.authCfg.WaitTimeMinutesOnFailedLogin15Times)
			afterLastAttempt := lastAttempt.Add(cooldown)
			if now.Before(afterLastAttempt) {
				duration := afterLastAttempt.Sub(now)
				timeUntil = now.Add(duration)
			}
		}

		if counter >= consecLoginAttempt3 {
			return models.GetTimeUntilLoginUnlockResponse{
				Until: timeUntilLoginIsPermanent,
			}, errMemberBlockedOnTooManyLoginAttempt
		}

		return models.GetTimeUntilLoginUnlockResponse{
			Until: timeUntil.Format(time.RFC3339),
		}, nil
	}

	return models.GetTimeUntilLoginUnlockResponse{
		Until: timeUntilLoginIsUnlocked,
	}, nil
}

func (svc *authService) TokenFromRefreshToken(ctx context.Context, req models.TokenFromRefreshTokenRequest) (models.AuthToken, error) {
	return svc.authTokenFromRefreshToken(ctx, req.RefreshToken)
}

func (svc *authService) authTokenFromRefreshToken(ctx context.Context, refreshToken string) (models.AuthToken, error) {
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		return svc.authCfg.JWT.SigningKey.Decoded, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed), errors.Is(err, jwt.ErrTokenUnverifiable),
			errors.Is(err, jwt.ErrTokenSignatureInvalid), errors.Is(err, jwt.ErrTokenExpired),
			errors.Is(err, jwt.ErrTokenUsedBeforeIssued):
			return models.AuthToken{}, errors.New(customerror.ErrGroupClientErr, "1", "Invalid refresh token")
		}

		return models.AuthToken{}, customerror.ErrInternal.WithError(err).WithSource()
	}

	cardNumCipherBase64, err := token.Claims.GetSubject()
	if err != nil {
		return models.AuthToken{}, customerror.ErrInternal.WithError(err).WithSource()
	}

	cardNumCipher, err := base64.RawStdEncoding.DecodeString(cardNumCipherBase64)
	if err != nil {
		return models.AuthToken{}, customerror.ErrInternal.WithError(err).WithSource()
	}

	cardNumB, err := encryption.AES256Decrypt(cardNumCipher, svc.authCfg.CardNumberEncryptionKey.Decoded)
	if err != nil {
		return models.AuthToken{}, customerror.ErrInternal.WithError(err).WithSource()
	}

	// Check if member has any mutation
	membership, err := svc.memberRepo.GetMembershipStatus(ctx, string(cardNumB))
	if err != nil {
		if err == customerror.ErrNotFound {
			return models.AuthToken{}, customerror.ErrMemberNotFound
		}

		return models.AuthToken{}, err
	}

	switch membership.Status {
	case memberMutationStatusWasMissing:
		return models.AuthToken{}, errMemberWasMissing
	case memberMutationStatusWasUpgraded:
		return models.AuthToken{}, errMemberWasUpgraded
	}

	// Check if member permanently locked from login
	loginAttempts, loginAttempted, err := svc.memberRepo.GetFailedLoginAttempt(ctx, membership.MemberID, membership.MemberRegistrationBranchID)
	if err != nil {
		return models.AuthToken{}, err
	}

	if loginAttempted {
		// We increment the counter by 1 because somehow the counter
		// starts from 0, not 1
		counter := loginAttempts.Counter + 1
		if counter >= consecLoginAttempt3 {
			return models.AuthToken{}, errMemberBlockedOnTooManyLoginAttempt
		}
	}

	// Generate new auth token
	authToken, err := svc.createAuthTokenForMember(membership.MemberID, membership.MemberRegistrationBranchID, string(cardNumB))

	return authToken, err
}

func (svc *authService) createJwt(scope string, subject string, expiresAtFromNow time.Duration) (string, time.Time, error) {
	now := time.Now()
	jti, err := uuid.NewV7()
	if err != nil {
		return "", time.Time{}, customerror.ErrInternal.WithError(err).WithSource()
	}

	expiresAt := now.Add(expiresAtFromNow)
	claims := entities.JWTClaims{
		Scope: scope,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti.String(),
			Subject:   subject,
			Issuer:    svc.authCfg.JWT.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			Audience:  []string{svc.authCfg.JWT.Audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(svc.authCfg.JWT.SigningKey.Decoded)
	if err != nil {
		return "", time.Time{}, customerror.ErrInternal.WithError(err).WithSource()
	}

	return tokenStr, expiresAt, nil
}

func (*authService) memberIdAndRegistBranchIdAsJwtSub(memberId, memberRegistBranchId int) string {
	return fmt.Sprintf("%d-%d", memberId, memberRegistBranchId)
}

func (svc *authService) RegisterEmailByOTP(ctx context.Context, memberId, memberRegBranchId int, email string) (models.OTPDetail, error) {
	// TODO: For later improvement, we can use random salt for the password so that identical passwords
	// will not have the same hash

	member, err := svc.memberRepo.GetMember(ctx, memberId, memberRegBranchId)
	if err != nil {
		return models.OTPDetail{}, customerror.ErrInternal.WithError(err).WithSource()
	}

	// Check if email is already taken
	emailExists, err := svc.memberRepo.IsEmailRegistered(ctx, email)
	if err != nil {
		return models.OTPDetail{}, err
	}

	if emailExists {
		return models.OTPDetail{}, errors.New(customerror.ErrGroupClientErr, "", "Email telah terdaftar sebelumnya")
	}

	now := time.Now()

	oldOtp, err := svc.regOtpRepo.Get(ctx, memberId, memberRegBranchId)
	if err != nil {
		if err != customerror.ErrNotFound {
			return models.OTPDetail{}, err
		}
	}

	if err == nil {
		if now.Before(oldOtp.NextRegeneration.Time()) {
			return models.OTPDetail{}, errors.New(customerror.ErrGroupClientErr, "", "Tunggu beberapa saat lagi untuk mengirim OTP")
		}
	}

	otp := random.GetRandomNumerics(svc.authCfg.OTP.Length)
	expiredAt := now.Add(time.Duration(svc.authCfg.OTP.ExpireDuration))
	nextRegen := now.Add(time.Duration(svc.authCfg.OTP.RegenerationCooldownDuration))
	err = svc.regOtpRepo.Save(ctx, entities.RegistrationOTP{
		MemberID:          memberId,
		MemberRegBranchID: memberRegBranchId,
		OTP:               otp,
		ExpiredAt:         sqltypes.NewTimestamp(expiredAt),
		NextRegeneration:  sqltypes.NewTimestamp(nextRegen),
	})

	if err != nil {
		return models.OTPDetail{}, err
	}

	// TODO: This is not the best practice, but for now this will do.
	// If you do have the chance to improve this, please do so
	durationInt := time.Duration(svc.authCfg.OTP.ExpireDuration) / time.Minute

	svc.emailSvc.Send(ctx, emailModule.Message{
		To:           []string{email},
		Subject:      "Verifikasi email",
		TemplateName: constants.EmailTemplateEmailVerificationOTP,
		TemplateData: emailModule.EmailVerificationOTPTemplateData{
			MemberName:      member.Name,
			OTP:             otp,
			DurationDisplay: fmt.Sprintf("%d menit", durationInt),
			CopyrightYear:   strconv.Itoa(time.Now().Year()),
		},
	})

	return models.OTPDetail{
		Digits:           svc.authCfg.OTP.Length,
		ExpiredAt:        expiredAt,
		NextRegeneration: nextRegen,
	}, nil
}

func (svc *authService) ValidateOTPForEmailAndPasswordRegistration(ctx context.Context, memberId, memberRegBranchId int, req models.ValidateOTPForEmailAndPasswordRegistrationRequest) error {
	now := time.Now()

	// Check if email is already taken
	emailExists, err := svc.memberRepo.IsEmailRegistered(ctx, req.Email)
	if err != nil {
		return err
	}

	if emailExists {
		return errors.New(customerror.ErrGroupClientErr, "", "Email telah terdaftar sebelumnya")
	}

	regOtp, err := svc.regOtpRepo.Get(ctx, memberId, memberRegBranchId)
	if err != nil {
		return customerror.ErrInternal.WithError(err).WithSource()
	}

	if regOtp.OTP != req.OTP {
		return errInvalidOtp
	}

	if now.After(regOtp.ExpiredAt.Time()) {
		return errInvalidOtp
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return customerror.ErrInternal.WithError(err).WithSource()
	}

	err = svc.memberRepo.SaveMemberEmailAndPassword(ctx, memberId, memberRegBranchId, req.Email, hashedPass)
	if err != nil {
		return err
	}

	if err := svc.regOtpRepo.Delete(ctx, memberId, memberRegBranchId); err != nil {
		return err
	}

	return nil
}

func (svc *authService) RegisterEmailFromGoogleSignin(ctx context.Context, memberId, memberRegBranchId int, idTokenStr string) (models.AuthToken, error) {
	member, err := svc.memberRepo.GetMember(ctx, memberId, memberRegBranchId)
	if err != nil {
		return models.AuthToken{}, customerror.ErrInternal.WithError(err).WithSource()
	}

	payload, err := idtoken.Validate(ctx, idTokenStr, svc.authCfg.Google.ClientID)
	if err != nil {
		return models.AuthToken{}, errors.New(customerror.ErrGroupUnauthorized, "", "invalid OAuth token")
	}

	// TODO: It is probably better to store this claims keys in constants, or better yet create helper function for it, but for now this will work
	googleUserId, _ := payload.Claims["sub"].(string)
	email, _ := payload.Claims["email"].(string)

	err = svc.memberRepo.SetMemberEmailAndGoogleUserID(ctx, memberId, memberRegBranchId, email, googleUserId)
	if err != nil {
		return models.AuthToken{}, err
	}

	return svc.createAuthTokenForMember(memberId, memberRegBranchId, member.CardNumber)
}

func (svc *authService) LoginWithEmailAndPassword(ctx context.Context, memberId, memberRegBranchId int, req models.LoginWithEmailAndPasswordRequest) (models.LoginResponse, error) {
	// TODO: Just like login with member card number and phone number,
	// we also need to have retry limit and cooldown to prevent brute force attack
	creds, err := svc.memberRepo.GetMemberEmailCreds(ctx, memberId, memberRegBranchId)
	if err != nil {
		if err == customerror.ErrNotFound {
			return models.LoginResponse{}, errLoginInvalidCreds
		}

		return models.LoginResponse{}, err
	}

	// Check if user registered email from Google Sign-in
	if creds.GoogleAccountID != "" {
		// TODO: Maybe ask about the wording of the error to design team, its probably will be used/shown to the user
		return models.LoginResponse{}, errors.New(customerror.ErrGroupUnauthorized, "", "Email didaftarkan melalui Google Sign-in, silahkan login kembali dengan Google Sign-in")
	}

	switch creds.MutationStatus {
	case memberMutationStatusWasMissing:
		return models.LoginResponse{}, errMemberWasMissing
	case memberMutationStatusWasUpgraded:
		return models.LoginResponse{}, errMemberWasUpgraded
	}

	if req.Email != creds.Email {
		return models.LoginResponse{}, errLoginInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword(creds.Password, []byte(req.Password)); err != nil {
		return models.LoginResponse{}, errLoginInvalidCreds
	}

	authToken, err := svc.createAuthTokenForMember(creds.MemberID, creds.MemberRegistrationBranchID, creds.CardNumber)
	if err != nil {
		return models.LoginResponse{}, err
	}

	return models.LoginResponse{
		AuthToken: authToken,
	}, nil
}

func (svc *authService) LoginWithGoogleSignin(ctx context.Context, memberId, memberRegBranchId int, idTokenStr string) (models.LoginResponse, error) {
	// TODO: Just like login with member card number and phone number,
	// we also need to have retry limit and cooldown to prevent brute force attack
	creds, err := svc.memberRepo.GetMemberEmailCreds(ctx, memberId, memberRegBranchId)
	if err != nil {
		if err == customerror.ErrNotFound {
			return models.LoginResponse{}, errLoginInvalidCreds
		}

		return models.LoginResponse{}, err
	}

	if creds.GoogleAccountID == "" {
		return models.LoginResponse{}, errors.New(customerror.ErrGroupUnauthorized, "", "Email tidak didaftarkan melalui Google Sign-in, silahkan login kembali menggunakan email dan password")
	}

	_, err = idtoken.Validate(ctx, idTokenStr, svc.authCfg.Google.ClientID)
	if err != nil {
		return models.LoginResponse{}, errors.New(customerror.ErrGroupUnauthorized, "", "invalid OAuth token")
	}

	authToken, err := svc.createAuthTokenForMember(memberId, memberRegBranchId, creds.CardNumber)
	if err != nil {
		return models.LoginResponse{}, err
	}

	return models.LoginResponse{
		AuthToken: authToken,
	}, nil
}
