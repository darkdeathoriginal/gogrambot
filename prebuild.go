//go:build prebuild
// +build prebuild

package main

import (
	"github.com/darkdeathoriginal/gogrambot/helpers"
	"github.com/darkdeathoriginal/gogrambot/models"
)

func main() {
	models.InitDatabase()
	helpers.LoadExternalPlugins()
}
