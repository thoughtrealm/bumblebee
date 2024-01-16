// Copyright 2023 The Bumblebee Authors
//
// Use of this source code is governed by an MIT license that is located
// in this project's root folder, and can also be found online at:
//
// https://github.com/thoughtrealm/bumblebee/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helpers

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/kirsle/configdir"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

var GlobalUseProfile string

const (
	BBGLobalFolderName = "Bumblebee"
	BBConfigFileName   = "config.yaml"
)

var GlobalConfig *ConfigHelper

type Profile struct {
	Name                  string `yaml:"name"`
	Path                  string `yaml:"path"`
	KeyStorePath          string `yaml:"keyStore"`
	KeyPairStorePath      string `yaml:"keyPairStore"`
	KeyPairStoreEncrypted bool   `yaml:"keyPairStoreEncrypted"`

	// DefaultKeypairName is optional and is the name to use as the sender when using the default key for this profile
	DefaultKeypairName string `yaml:"defaultKeypairName"`
}

func (p *Profile) Clone() *Profile {
	return &Profile{
		Name:                  p.Name,
		Path:                  p.Path,
		KeyStorePath:          p.KeyStorePath,
		KeyPairStorePath:      p.KeyPairStorePath,
		KeyPairStoreEncrypted: p.KeyPairStoreEncrypted,
		DefaultKeypairName:    p.DefaultKeypairName,
	}
}

type ConfigInfo struct {
	Profiles       []*Profile `yaml:"profiles"`
	CurrentProfile string     `yaml:"currentProfile"`
}

type YAMLConfig struct {
	Config *ConfigInfo `yaml:"BumblebeeSettings"`
}

type ConfigHelper struct {
	Config *ConfigInfo
}

func NewConfigHelper() *ConfigHelper {
	return &ConfigHelper{}
}

func NewConfigHelperFromConfig(config *ConfigInfo) *ConfigHelper {
	nc := NewConfigHelper()
	nc.Config = config.Clone()
	return nc
}

func (configIn *ConfigInfo) Clone() (configOut *ConfigInfo) {
	configOut = &ConfigInfo{CurrentProfile: configIn.CurrentProfile}
	for _, profile := range configIn.Profiles {
		newProfile := &Profile{
			Name:                  profile.Name,
			Path:                  profile.Path,
			KeyStorePath:          profile.KeyStorePath,
			KeyPairStorePath:      profile.KeyPairStorePath,
			KeyPairStoreEncrypted: profile.KeyPairStoreEncrypted,
			DefaultKeypairName:    profile.DefaultKeypairName,
		}

		configOut.Profiles = append(configOut.Profiles, newProfile)
	}

	return configOut
}

func GetConfigPath() (string, error) {
	configPath := configdir.LocalConfig(BBGLobalFolderName)
	err := configdir.MakePath(configPath)
	if err != nil {
		return "", fmt.Errorf("failed validating existance of config paths: %w", err)
	}

	return configPath, nil
}

func BuildProfilePath(name string) (string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return "", err
	}

	profilePath := filepath.Join(configPath, name)
	err = configdir.MakePath(profilePath)
	if err != nil {
		return "", fmt.Errorf("failed validating profile path: %w", err)
	}

	return profilePath, nil
}

func (ch *ConfigHelper) LoadConfig() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	configFilePath := filepath.Join(configPath, BBConfigFileName)
	configBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("failed config file: %w", err)
	}

	var configYaml YAMLConfig
	err = yaml.Unmarshal(configBytes, &configYaml)
	if err != nil {
		return fmt.Errorf("failed interpreting config file data: %w", err)
	}

	ch.Config = configYaml.Config
	return nil
}

func (ch *ConfigHelper) WriteConfig() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	configFilePath := filepath.Join(configPath, BBConfigFileName)
	configYaml := YAMLConfig{Config: ch.Config}

	configBytes, err := yaml.Marshal(configYaml)
	if err != nil {
		return fmt.Errorf("failed marshaling configYaml: %w", err)
	}

	err = os.WriteFile(configFilePath, configBytes, 0666)
	if err != nil {
		return fmt.Errorf("failed writing config file: %w", err)
	}

	return nil
}

func (ch *ConfigHelper) GetConfigYAMLLines() ([]string, error) {
	if ch.Config == nil {
		return nil, errors.New("unable to build YAML Lines: config not loaded")
	}

	configYaml := YAMLConfig{Config: ch.Config}

	yamlBytes, err := yaml.Marshal(configYaml)
	if err != nil {
		return nil, fmt.Errorf("failed reading data: %w", err)
	}

	yamlLinesBytes := bytes.Split(yamlBytes, []uint8("\n"))
	yamlLines := make([]string, len(yamlLinesBytes))
	for i, lineBytes := range yamlLinesBytes {
		yamlLines[i] = string(lineBytes)
	}

	return yamlLines, nil
}

func (ch *ConfigHelper) GetProfileYAMLLines(profileName string) ([]string, string, error) {
	if ch.Config == nil {
		return nil, "", fmt.Errorf("unable to build YAML Lines: %s", errors.New("config not loaded"))
	}

	if profileName == "" {
		if GlobalUseProfile == "" {
			profileName = GlobalConfig.Config.CurrentProfile
		} else {
			profileName = GlobalUseProfile
		}

	}

	if profileName == "" {
		return nil, "", fmt.Errorf("Unable to build YAML lines: %s", errors.New("no profile name provided or found in config"))
	}

	profile := GlobalConfig.GetProfile(profileName)
	if profile == nil {
		return nil, profileName, fmt.Errorf("Unable to build YAML Lines: %w", fmt.Errorf("Profile name not found in config: %s", profileName))

	}

	yamlBytes, err := yaml.Marshal(profile)
	if err != nil {
		return nil, profileName, fmt.Errorf("failed reading profile data: %w", err)
	}

	yamlLinesBytes := bytes.Split(yamlBytes, []uint8("\n"))
	yamlLines := make([]string, len(yamlLinesBytes))
	for i, lineBytes := range yamlLinesBytes {
		yamlLines[i] = string(lineBytes)
	}

	return yamlLines, profileName, nil
}

func (ch *ConfigHelper) GetConfigInfo() *ConfigInfo {
	return ch.Config.Clone()
}

func (ch *ConfigHelper) SelectProfile(name string) error {
	return errors.New("not implemented")
}

func (ch *ConfigHelper) ListProfiles() []Profile {
	return nil
}

func (ch *ConfigHelper) RemoveProfile(name string) (found bool, err error) {
	if ch == nil {
		return false, errors.New("config not initialized")
	}

	if ch.Config == nil {
		return false, errors.New("config not loaded")
	}

	nameUpper := strings.ToUpper(name)
	foundIdx := -1
	for idx, profile := range ch.Config.Profiles {
		if strings.ToUpper(profile.Name) == nameUpper {
			foundIdx = idx
			break
		}
	}

	if foundIdx == -1 {
		// we did not find the referenced profile, so just return
		return false, nil
	}

	if len(ch.Config.Profiles) == 1 {
		// Only one profile, so reset the profiles slice
		ch.Config.Profiles = make([]*Profile, 0)
	} else {
		// more than one, so remove the specific profile entry in the profiles slice
		ch.Config.Profiles = append(ch.Config.Profiles[:foundIdx], ch.Config.Profiles[foundIdx+1:]...)
	}

	// Now, make sure the default profile is not the one we removed
	if strings.ToUpper(ch.Config.CurrentProfile) == nameUpper {
		if len(ch.Config.Profiles) == 0 {
			ch.Config.CurrentProfile = ""
		} else {
			ch.Config.CurrentProfile = ch.Config.Profiles[0].Name
		}
	}

	return true, nil
}

func (ch *ConfigHelper) RenameProfile(currentName, newName string) (found bool, err error) {
	return false, errors.New("not implemented")
}

func (ch *ConfigHelper) GetCurrentProfile() *Profile {
	var currentProfileName string
	if GlobalUseProfile == "" {
		currentProfileName = strings.ToUpper(ch.Config.CurrentProfile)
	} else {
		currentProfileName = strings.ToUpper(GlobalUseProfile)
	}

	for _, profile := range ch.Config.Profiles {
		if strings.ToUpper(profile.Name) == currentProfileName {
			return profile
		}
	}

	return nil
}

func (ch *ConfigHelper) GetProfile(name string) *Profile {
	nameUpper := strings.ToUpper(name)
	for _, profile := range ch.Config.Profiles {
		if strings.ToUpper(profile.Name) == strings.ToUpper(nameUpper) {
			return profile
		}
	}

	return nil
}

func (ch *ConfigHelper) NewProfile(profile *Profile) error {
	if ch.Config == nil {
		return fmt.Errorf("unable to add new profile: %w", errors.New("config not currently loaded"))
	}

	if profile.Name == "" {
		return fmt.Errorf("unable to add new profile: %w", errors.New("supplied profile name is empty"))
	}

	// make sure a profile does not currently exist with this name
	profileCheck := ch.GetProfile(profile.Name)
	if profileCheck != nil {
		return fmt.Errorf("unable to add new profile: %w", errors.New("a profile currently exists with this name"))
	}

	ch.Config.Profiles = append(ch.Config.Profiles, profile.Clone())

	return nil
}
