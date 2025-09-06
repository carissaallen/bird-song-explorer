package config

import (
	"time"
)

// VoiceProfile represents a vetted voice configuration
type VoiceProfile struct {
	ID       string
	Name     string
	Region   string
	Language string
}

// VoiceManager handles voice selection and consistency
type VoiceManager struct {
	availableVoices []VoiceProfile
	dailyVoice      *VoiceProfile
	dailySeed       int
}

// Voices chosen from different regions
// NOTE: Run scripts/generate_regional_intros.sh to create intros for new voices
var DefaultVoices = []VoiceProfile{
	{
		ID:       "ZF6FPAbjXT4488VcRRnw",
		Name:     "Amelia",
		Region:   "British",
		Language: "en-GB",
	},
	{
		ID:       "ErXwobaYiN019PkySvjV",
		Name:     "Antoni",
		Region:   "American",
		Language: "en-US",
	},
	{
		ID:       "HDA9tsk27wYi3uq0fPcK",
		Name:     "Stuart",
		Region:   "Australian",
		Language: "en-AU",
	},
	{
		ID:       "FVQMzxJGPUBtfz1Azdoy",
		Name:     "Danielle",
		Region:   "Canadian",
		Language: "en-CA",
	},
	//{
	//	ID:       "1SM7GgM6IMuvQlz2BwM3",
	//	Name:     "Mark",
	//	Region:   "American",
	//	Language: "en-US",
	//},
	//{
	//	ID:       "5GZaeOOG7yqLdoTRsaa6",
	//	Name:     "Charlotte",
	//	Region:   "Australian",
	//	Language: "en-AU",
	//},
	//{
	//	ID:       "iCrDUkL56s3C8sCRl7wb",
	//	Name:     "Hope",
	//	Region:   "American",
	//	Language: "en-US",
	//},
	//{
	//	ID:       "hmMWXCj9K7N5mCPcRkfC",
	//	Name:     "Rory",
	//	Region:   "Irish",
	//	Language: "en-IE",
	//},
}

// NewVoiceManager creates a new voice manager with default voices
func NewVoiceManager() *VoiceManager {
	return &VoiceManager{
		availableVoices: DefaultVoices,
	}
}

// GetDailyVoice returns the voice for the current day
// This ensures all tracks generated on the same day use the same voice
func (vm *VoiceManager) GetDailyVoice() VoiceProfile {
	now := time.Now()
	seed := now.Year()*10000 + int(now.Month())*100 + now.Day()

	// If it's a new day or first call, select a voice
	if vm.dailyVoice == nil || vm.dailySeed != seed {
		vm.dailySeed = seed
		// Use modulo to ensure consistent selection across instances
		// This gives us deterministic selection based on the day
		voiceIndex := seed % len(vm.availableVoices)
		vm.dailyVoice = &vm.availableVoices[voiceIndex]
	}

	return *vm.dailyVoice
}

// GetVoiceByID returns a specific voice by ID if it exists
func (vm *VoiceManager) GetVoiceByID(voiceID string) *VoiceProfile {
	for _, voice := range vm.availableVoices {
		if voice.ID == voiceID {
			return &voice
		}
	}
	return nil
}

// GetVoiceByName returns a specific voice by name if it exists
func (vm *VoiceManager) GetVoiceByName(name string) *VoiceProfile {
	for _, voice := range vm.availableVoices {
		if voice.Name == name {
			return &voice
		}
	}
	return nil
}

// GetAvailableVoices returns all configured voices
func (vm *VoiceManager) GetAvailableVoices() []VoiceProfile {
	return vm.availableVoices
}

// SetAvailableVoices updates the list of available voices
func (vm *VoiceManager) SetAvailableVoices(voices []VoiceProfile) {
	vm.availableVoices = voices
	// Reset daily voice to force reselection
	vm.dailyVoice = nil
}
