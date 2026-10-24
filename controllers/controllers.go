package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qutty/pushserver/config"
	"github.com/qutty/pushserver/models"
	"github.com/qutty/pushserver/providers"
	"gorm.io/gorm"
)

var (
	db            *gorm.DB
	providerCache sync.Map
)

func SetDB(database *gorm.DB) {
	db = database
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GET /v1/providers
func ListProviders(c *gin.Context) {
	provs := make([]string, 0)
	if config.AppConfig != nil {
		for k := range config.AppConfig.Providers {
			provs = append(provs, k)
		}
	}
	c.JSON(http.StatusOK, gin.H{"providers": provs})
}

// POST /v1/reg
func Register(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	provName, _ := req["provider"].(string)
	if provName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider required"})
		return
	}

	var prov providers.Provider
	if cached, ok := providerCache.Load(provName); ok {
		prov = cached.(providers.Provider)
	} else {
		if config.AppConfig != nil {
			if pconf, ok := config.AppConfig.Providers[provName]; ok {
				if pconf.Type == "fcm" {
					prov = providers.NewFCMProvider(pconf.Arguments)
				} else if pconf.Type == "apple" {
					prov = providers.NewAppleProvider(pconf.Arguments)
				} else {
					base := provName
					prov = providers.NewBaseProvider(base)
				}
			} else {
				base := provName
				prov = providers.NewBaseProvider(base)
			}
		} else {
			base := provName
			prov = providers.NewBaseProvider(base)
		}
		providerCache.Store(provName, prov)
	}

	token := generateToken()

	if db != nil {
		record := models.TokenRecord{
			Token:    token,
			Provider: provName,
			Active:   true,
		}
		db.Create(&record)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "token": token})
}

// POST /v1/unreg
func Unregister(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	token, _ := req["token"].(string)
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
		return
	}
	if db != nil {
		db.Where("token = ?", token).Delete(&models.TokenRecord{})
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// POST /v1/alive
func Alive(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	token, _ := req["token"].(string)
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
		return
	}
	if db != nil {
		db.Model(&models.TokenRecord{}).Where("token = ?", token).Update("updated_at", time.Now())
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
