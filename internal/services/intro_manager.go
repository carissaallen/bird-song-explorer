package services

import (
	"fmt"
	"math/rand"
	"time"
)

type IntroManager struct {
	intros []string
}

func NewIntroManager() *IntroManager {
	return &IntroManager{
		intros: []string{
			"Welcome, nature detectives! Time to discover an amazing bird from your neighborhood.",
			"Hello, bird explorers! Today's special bird is waiting to sing for you.",
			"Ready for an adventure? Let's meet today's featured bird from your area!",
			"Welcome back, little listeners! A wonderful bird is calling just for you.",
			"Hello, young scientists! Let's explore the amazing birds living near you.",
			"Calling all bird lovers! Your daily bird discovery is ready.",
			"Time for today's bird adventure! Listen closely to nature's music.",
			"Welcome to your daily bird journey! Let's discover who's singing today.",
		},
	}
}

func (im *IntroManager) GetRandomIntro() string {
	rand.Seed(time.Now().UnixNano())
	return im.intros[rand.Intn(len(im.intros))]
}

func (im *IntroManager) GetIntroForBird(birdName string) string {
	templates := []string{
		"Today's featured friend is the %s! Let's hear their beautiful song.",
		"Listen closely! The amazing %s has something special to share with you.",
		"Get ready to meet the wonderful %s from your neighborhood!",
		"Your bird discovery today is the %s! What an incredible creature!",
	}

	rand.Seed(time.Now().UnixNano())
	template := templates[rand.Intn(len(templates))]
	return fmt.Sprintf(template, birdName)
}
