package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var dbPtr *gorm.DB

var logger *slog.Logger

type device struct {
	DeviceID    string    `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	OS          string    `json:"os"`
	CreatedTime time.Time `json:"created_time"`
	UpdatedTime time.Time `json:"updated_time"`
	Owner       string    `json:"owner"`
	Status      string    `json:"status"`
}

var deviceMap map[string]device

// const dbName = "device-postgres"
func createDatabase() {
	defaultDBStr := "host=postgres user=postgres password=secret dbname=postgres port=5432 sslmode=disable"
	var err error

	maxRetries := 10
	retryDelay := 2 * time.Second

	logfile, err := os.OpenFile("device.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("failed to open log file: %v", err)
	}

	logger = slog.New(slog.NewJSONHandler(logfile, &slog.HandlerOptions{
		Level: slog.LevelDebug, AddSource: true,
	}))
	slog.SetDefault(logger)

	for i := 0; i < maxRetries; i++ {
		dbPtr, err = gorm.Open(postgres.Open(defaultDBStr), &gorm.Config{})
		if err == nil {
			slog.Info("Successfully connected to the database.")
			if err := dbPtr.AutoMigrate(&device{}); err != nil {
				slog.Error("Failed to migrate the DB", "Error", err)
			}
			break
		}
		msg := fmt.Sprintf("Database connection failed attempt %d/%d \n", i+1, maxRetries)
		slog.Error(msg, "Error", err)
		time.Sleep(retryDelay)
	}
	// log.Fatalf("Could not connect to the database after %d attempts: %v", maxRetries, err)
	// fmt.Println("DB Pointer: ", dbPtr)
}

func main() {
	createDatabase()
	router := gin.Default()
	router.GET("/v1/devices", getDevices)
	router.POST("/v1/devices", postDevices)
	router.GET("/v1/devices/:id", getDeviceByID)
	router.GET("/v1/devices/:id/status", getStatusByDeviceID)
	router.PATCH("/v1/devices", patchDeviceInfo)
	router.POST("/v1/loglevel/:level")

	deviceMap = make(map[string]device)
	for _, device := range devices {
		deviceMap[device.DeviceID] = device
	}

	router.Run(":8080")
}

var devices = []device{
	{DeviceID: "1", Name: "MacBook M1", Type: "Laptop", OS: "MacOS", Owner: "Guru", Status: "Active"},
	{DeviceID: "2", Name: "Lenovo LOQ", Type: "Laptop", OS: "Windows", Owner: "Nithin", Status: "Inactive"},
	{DeviceID: "3", Name: "MacBook M2", Type: "Laptop", OS: "MacOS", Owner: "Porkodi", Status: "Active"},
}

func getDevices(ctx *gin.Context) {
	slog.Debug("Entering getDevices API Call")
	var device []device

	result := dbPtr.Find(&device)

	if result.Error != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"error": result.Error})
		slog.Error("Error Retrieving Device Details", "error", result.Error)
		return
	}

	slog.Info("Device List Retrieved Successfully!", "Number of devices", len(device))
	ctx.IndentedJSON(http.StatusOK, device)
}

func getDeviceByID(ctx *gin.Context) {
	slog.Debug("Entering getDeviceByID API Call")
	id := ctx.Param("id")
	var device device

	if err := dbPtr.First(&device, id).Error; err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"error": "Invalid Device ID"})
		slog.Error("Error Retrieving Device Details by ID", "error", err)
		return
	}

	slog.Info("Device Details Retrieved Successfully!", "Device ID", device.DeviceID, "Device name", device.Name)
	ctx.IndentedJSON(http.StatusOK, device)
}

func getStatusByDeviceID(ctx *gin.Context) {
	slog.Debug("Entering getStatusByDeviceID API Call")
	id := ctx.Param("id")
	var device device

	if err := dbPtr.First(&device, id).Error; err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"error": "Invalid Device ID"})
		slog.Error("Error Retrieving Device Status from ID", "error", err)
		return
	}

	slog.Info("Device Status Retrieved Successfully!", "Device ID", device.DeviceID, "Device status", device.Status)

	ctx.JSON(http.StatusOK, gin.H{
		"Device ID":     device.DeviceID,
		"Device Status": device.Status,
	})

	// device, exists := deviceMap[id]
	// if exists {
	// 	ctx.IndentedJSON(http.StatusOK, device.Status)
	// 	return
	// }

	// ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Device ID"})
}

func postDevices(ctx *gin.Context) {
	slog.Debug("Entering postDevices API Call")
	var newDevice device

	if err := ctx.ShouldBindJSON(&newDevice); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Error("Error Mapping Device Structure to JSON", "error", err)
		return
	}

	if err := dbPtr.Create(&newDevice).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not register device"})
		slog.Error("Error Creating Device Details", "error", err)
		return
	}

	// _, exists := deviceMap[newDevice.ID]
	// if exists {
	// 	ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Device ID already exists"})
	// 	return
	// }

	// deviceMap[newDevice.ID] = newDevice
	// devices = append(devices, newDevice)

	slog.Info("New Device Successfully Created!", "Device ID", newDevice.DeviceID, "Device name", newDevice.Name)
	ctx.IndentedJSON(http.StatusCreated, newDevice)
}

func patchDeviceInfo(ctx *gin.Context) {
	slog.Debug("Entering patchDeviceInfo API Call")
	id := ctx.Param("id")
	var devices device

	if err := dbPtr.First(&devices, id).Error; err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		slog.Error("Error Updating Device Details", "error", err)
		return
	}

	var updateData device

	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Patch Request"})
		return
	}

	updateData.UpdatedTime = time.Now()

	if err := dbPtr.Save(&updateData).Error; err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// if err := ctx.ShouldBindJSON(&deviceInfo); err != nil {
	// 	ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	// _, exists := deviceMap[deviceInfo.DeviceID]
	// if !exists {
	// 	ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Device does not exist"})
	// 	return
	// }

	// deviceMap[deviceInfo.DeviceID] = deviceInfo

	slog.Info("Device Details Retrieved Successfully!", "Device ID", devices.DeviceID, "Name", devices.Name)
	ctx.IndentedJSON(http.StatusOK, gin.H{"success": "Device Object Patched Successfully"})
}

// 	r.GET("/api/devices", listDevices)
// r.POST("/api/devices", registerDevice)
// 	r.GET("/api/devices/:id", getDevice)
// 	r.PUT("/api/devices/:id", updateDevice)
// 	r.GET("/api/devices/:id/status", monitorDevice)

// 	r.POST("/api/devices/:id/commands", sendCommand)

// debug, info level
// new rest api for log level setting and getting
// lumberack for log rotation
