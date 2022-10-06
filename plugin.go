// MIT License
//
// Copyright (c) 2022 Spiral Scout
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package config

import (
	"github.com/roadrunner-server/errors"
	"github.com/rumorsflow/contracts/config"
	"github.com/spf13/viper"
	"os"
	"strings"
	"time"
)

var _ config.Configurer = (*Plugin)(nil)

const PluginName string = "config"

type Plugin struct {
	Path    string
	Prefix  string
	Version string
	Cmd     string
	Timeout time.Duration
	viper   *viper.Viper
}

// Init config provider.
func (p *Plugin) Init() error { //nolint:gocognit,gocyclo
	const op = errors.Op("config plugin init")

	p.viper = viper.New()
	p.viper.AutomaticEnv()
	p.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if p.Prefix == "" {
		return errors.E(op, errors.Init, errors.Str("prefix should be set"))
	}
	p.viper.SetEnvPrefix(p.Prefix)

	if p.Path == "" {
		return errors.E(op, errors.Init, errors.Str("path should be set"))
	}
	p.viper.SetConfigFile(p.Path)

	if err := p.viper.ReadInConfig(); err != nil {
		return errors.E(op, errors.Init, err)
	}

	// automatically inject ENV variables using ${ENV} pattern
	for _, key := range p.viper.AllKeys() {
		val := p.viper.Get(key)
		switch t := val.(type) {
		case string:
			// for string just expand it
			p.viper.Set(key, os.ExpandEnv(t))
		case []any:
			// for slice -> check if it's slice of strings
			strArr := make([]string, 0, len(t))
			for i := 0; i < len(t); i++ {
				if valStr, ok := t[i].(string); ok {
					strArr = append(strArr, os.ExpandEnv(valStr))
					continue
				}

				p.viper.Set(key, val)
			}

			// we should set the whole array
			if len(strArr) > 0 {
				p.viper.Set(key, strArr)
			}
		default:
			p.viper.Set(key, val)
		}
	}

	return nil
}

// UnmarshalKey reads configuration section into configuration object.
func (p *Plugin) UnmarshalKey(name string, out any) error {
	const op = errors.Op("config plugin unmarshal key")
	err := p.viper.UnmarshalKey(name, out)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (p *Plugin) Unmarshal(out any) error {
	const op = errors.Op("config plugin unmarshal")
	err := p.viper.Unmarshal(out)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// Overwrite overwrites existing config with provided values
func (p *Plugin) Overwrite(values map[string]any) error {
	for key, value := range values {
		p.viper.Set(key, value)
	}
	return nil
}

// Get raw config in a form of config section.
func (p *Plugin) Get(name string) any {
	return p.viper.Get(name)
}

// Has checks if config section exists.
func (p *Plugin) Has(name string) bool {
	return p.viper.IsSet(name)
}

// GetVersion returns app version
func (p *Plugin) GetVersion() string {
	return p.Version
}

// GetCmd returns cli command name
func (p *Plugin) GetCmd() string {
	return p.Cmd
}

// GracefulTimeout represents timeout for all servers registered in endure
func (p *Plugin) GracefulTimeout() time.Duration {
	return p.Timeout
}

// Name returns user-friendly plugin name
func (p *Plugin) Name() string {
	return PluginName
}
