package logger

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Run tests
	os.Exit(m.Run())
}

func TestLocalLogger(t *testing.T) {
	testDevLogger := ConfigureLogger("deskpass-logger-test", "", "development")
	testProdLogger := ConfigureLogger("deskpass-logger-test", "", "production")

	standardTestSuite(t, testDevLogger)
	standardTestSuite(t, testProdLogger)
}

func TestRemoteLogger(t *testing.T) {
	loggingURL := os.Getenv("PAPERTRAIL_LOGGING_URL")

	if loggingURL == "" {
		t.Error("Unable to test remote logging functionality: PAPERTRAIL_LOGGING_URL not set")
	}

	os.Setenv("REMOTE_LOG_ONLY", "1")

	testDevLogger := ConfigureLogger("deskpass-logger-test", loggingURL, "development")
	testProdLogger := ConfigureLogger("deskpass-logger-test", loggingURL, "production")

	standardTestSuite(t, testDevLogger)
	standardTestSuite(t, testProdLogger)
}

func standardTestSuite(t *testing.T, logger *Logger) {
	testError := fmt.Errorf("This is a test error")
	testMeta := map[string]interface{}{
		"testKey":  "testString",
		"testKey2": 10,
		"testKey3": 69.420,
		"testKey4": true,
	}

	logger.Debug("This is a test debug message without meta info", nil)
	logger.Debug("This is a test debug message with meta info", testMeta)

	logger.Info("This is a test info message without meta info", nil)
	logger.Info("This is a test info message with meta info", testMeta)

	logger.Warn("This is a test warn message without meta info", nil)
	logger.Warn("This is a test warn message with meta info", testMeta)

	logger.Error("This is a test err message without error or meta info", nil, nil)
	logger.Error("This is a test err message with error and no meta info", testError, nil)
	logger.Error("This is a test err message with meta info", nil, testMeta)
	logger.Error("This is a test err message with error and meta info", testError, testMeta)
}
