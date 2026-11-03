package cli

import (
	"fmt"
	"os"

	"github.com/qutty/pushserver/config"
	"github.com/qutty/pushserver/controllers"
	"github.com/qutty/pushserver/middleware"
	"github.com/qutty/pushserver/models"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Run(args []string) {
	if len(args) == 0 {
		help()
		return
	}
	switch args[0] {
	case "help":
		if len(args) == 1 {
			help()
		} else if len(args) == 2 && args[1] == "types" {
			helpTypes()
		} else if len(args) == 3 && args[1] == "type" {
			helpType(args[2])
		} else {
			help()
		}
	case "test":
		testConfig()
	case "serve":
		StartServer()
	default:
		help()
	}
}

func help() {
	fmt.Println("Qutty Push Server v1.0")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  help                Show this help")
	fmt.Println("  help types          List available types")
	fmt.Println("  help type <type>    Show example config for type")
	fmt.Println("  test                Test config")
	fmt.Println("  serve               Start server")
	fmt.Println("")
}

func helpTypes() {
	fmt.Println("Available types:")
	fmt.Println("  fcm")
	fmt.Println("  apple")
	fmt.Println("")
}

func helpType(t string) {
	switch t {
case "fcm":
		fmt.Println(`Example config for fcm:
{
    "label":"example label",
    "providers": {
        "exampleprovider": {
            "type" : "fcm",
            "arguments": {
                "project_id": "...",
                "service_account": "..."
            }
        }
    }
}`)
	case "apple":
		fmt.Println(`Example config for apple:
{
    "label":"example label",
    "providers": {
        "exampleprovider": {
            "type" : "apple",
            "arguments": {
                "cert": "...",
                "key": "...",
                "provider": "https://..."
            }
        }
    }
}`)
	default:
		fmt.Println("Unknown type")
	}
}

func testConfig() {
	if _, err := os.Stat("config.json"); err != nil {
		fmt.Println("config.json not found")
		return
	}
	if err := config.LoadConfig("config.json"); err != nil {
		fmt.Println("Config error:", err)
		return
	}
	fmt.Println("Config is valid")
}

func StartServer() {
	r := gin.Default()
	
	// Rate limiting
	r.Use(middleware.GlobalLimiter.Limit())
	
	// In-memory token store (pure Go, no SQLite dependency)
	tokenStore := models.NewStore()
	controllers.SetStore(tokenStore)
	
	// Load config if exists
	if _, err := os.Stat("config.json"); err == nil {
		config.LoadConfig("config.json")
	}
	
	// Routes
	r.GET("/health", controllers.Health)
	r.GET("/v1/providers", controllers.ListProviders)
	r.POST("/v1/reg", controllers.Register)
	r.POST("/v1/unreg", controllers.Unregister)
	r.POST("/v1/alive", controllers.Alive)
	
	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	r.Run(":8877")
}

func init() {
	// Print banner
	printBanner()
}
