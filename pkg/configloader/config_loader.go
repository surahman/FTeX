package configloader

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/validator"
)

// Load will load configurations stored in a file system into a configuration container struct.
func Load[T any](fs afero.Fs, cfg *T, filename, prefix, format string) (err error) {
	viper.SetFs(fs)
	viper.SetConfigName(filename)
	viper.SetConfigType(format)
	viper.AddConfigPath(constants.GetEtcDir())
	viper.AddConfigPath(constants.GetHomeDir())
	viper.AddConfigPath(constants.GetBaseDir())

	viper.SetEnvPrefix(prefix)
	viper.AutomaticEnv()

	if err = viper.ReadInConfig(); err != nil {
		return
	}

	if err = viper.Unmarshal(cfg); err != nil {
		return
	}

	if err = validator.ValidateStruct(cfg); err != nil {
		return
	}

	return
}
