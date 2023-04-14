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
  expirationDuration: 600
  refreshThreshold: 60
general:
  bcryptCost: 8
  cryptoSecret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"no_issuer": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  expirationDuration: 600
  refreshThreshold: 60
general:
  bcryptCost: 8
  cryptoSecret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"bcrypt_cost_below_4": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  issuer: MCQ Platform
  expirationDuration: 600
  refreshThreshold: 60
general:
  bcryptCost: 2
  cryptoSecret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"bcrypt_cost_above_31": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  issuer: MCQ Platform
  expirationDuration: 600
  refreshThreshold: 60
general:
  bcryptCost: 32
  cryptoSecret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"jwt_expiration_below_60s": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  issuer: MCQ Platform
  expirationDuration: 59
  refreshThreshold: 40
general:
  bcryptCost: 8
  cryptoSecret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"jwt_key_below_8": `
jwt:
  key: kYzJdnp
  issuer: MCQ Platform
  expirationDuration: 600
  refreshThreshold: 60
general:
  bcryptCost: 8
  cryptoSecret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"jwt_key_above_256": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9UkYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9UkYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9UkYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9UkYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5
  issuer: MCQ Platform
  expirationDuration: 600
  refreshThreshold: 60
general:
  bcryptCost: 8
  cryptoSecret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"low_refresh_threshold": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  issuer: MCQ Platform
  expirationDuration: 600
  refreshThreshold: 0
general:
  bcryptCost: 8
  cryptoSecret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"refresh_threshold_gt_expiration": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  issuer: MCQ Platform
  expirationDuration: 600
  refreshThreshold: 601
general:
  bcryptCost: 8
  cryptoSecret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$`,

		"crypto_key_too_short": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  issuer: MCQ Platform
  expirationDuration: 600
  refreshThreshold: 60
general:
  bcryptCost: 8
  cryptoSecret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z`,

		"crypto_key_too_long": `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  issuer: MCQ Platform
  expirationDuration: 600
  refreshThreshold: 60
general:
  bcryptCost: 8
  cryptoSecret: Xp2s5v8y/B?E(H+MbQeShVmYq3t6w9z$*`,
	}
}
