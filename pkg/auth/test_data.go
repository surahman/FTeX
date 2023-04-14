package auth

// configTestData will return a map of test data containing valid and invalid Authorization configs.
//
//nolint:lll
func configTestData() map[string]string {
	return map[string]string{

		"empty": ``,

		"valid": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  issuer: MCQ Platform
  expiration_duration: 600
  refresh_threshold: 60
general:
  bcrypt_cost: 8
  crypto_secret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"no_issuer": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  expiration_duration: 600
  refresh_threshold: 60
general:
  bcrypt_cost: 8
  crypto_secret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"bcrypt_cost_below_4": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  issuer: MCQ Platform
  expiration_duration: 600
  refresh_threshold: 60
general:
  bcrypt_cost: 2
  crypto_secret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"bcrypt_cost_above_31": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  issuer: MCQ Platform
  expiration_duration: 600
  refresh_threshold: 60
general:
  bcrypt_cost: 32
  crypto_secret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"jwt_expiration_below_60s": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  issuer: MCQ Platform
  expiration_duration: 59
  refresh_threshold: 40
general:
  bcrypt_cost: 8
  crypto_secret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"jwt_key_below_8": `
jwt:
  key: kYzJdnp
  issuer: MCQ Platform
  expiration_duration: 600
  refresh_threshold: 60
general:
  bcrypt_cost: 8
  crypto_secret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"jwt_key_above_256": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9UkYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9UkYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9UkYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9UkYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5
  issuer: MCQ Platform
  expiration_duration: 600
  refresh_threshold: 60
general:
  bcrypt_cost: 8
  crypto_secret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"low_refresh_threshold": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  issuer: MCQ Platform
  expiration_duration: 600
  refresh_threshold: 0
general:
  bcrypt_cost: 8
  crypto_secret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"refresh_threshold_gt_expiration": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  issuer: MCQ Platform
  expiration_duration: 600
  refresh_threshold: 601
general:
  bcrypt_cost: 8
  crypto_secret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"crypto_key_too_short": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  issuer: MCQ Platform
  expiration_duration: 600
  refresh_threshold: 60
general:
  bcrypt_cost: 8
  crypto_secret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z`,

		"crypto_key_too_long": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  issuer: MCQ Platform
  expiration_duration: 600
  refresh_threshold: 60
general:
  bcrypt_cost: 8
  crypto_secret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$*`,
	}
}
