package config

import (
	"os"
	"strings"
)

type Config struct {
	Port      string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	DBSSLMode string

	// SIAKAD Enterprise DB (gate, ref, pendaftaran)
	DBSiakadHost    string
	DBSiakadPort    string
	DBSiakadUser    string
	DBSiakadPass    string
	DBSiakadName    string
	DBSiakadSSLMode string
}

func (c *Config) IsDualDB() bool {
	return c.DBSiakadName != "" && c.DBSiakadName != c.DBName
}

func loadDotEnv(filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			// Strip inline comments if not quoted
			if !strings.HasPrefix(v, `"`) && !strings.HasPrefix(v, `'`) {
				if commentIdx := strings.Index(v, "#"); commentIdx != -1 {
					v = strings.TrimSpace(v[:commentIdx])
				}
			}
			v = strings.Trim(v, `"'`)
			if os.Getenv(k) == "" {
				os.Setenv(k, v)
			}
		}
	}
}

func LoadConfig() *Config {
	// Automatically load .env file if present
	loadDotEnv(".env")
	loadDotEnv("../.env")

	port := getEnv("PORT", "5000")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5433")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASS", "postgres")
	dbName := getEnv("DB_NAME", "cat")
	dbSSL := getEnv("DB_SSLMODE", "disable")

	// SIAKAD fallback to primary DB config if not specified
	dbSiakadHost := getEnv("DB_SIAKAD_HOST", dbHost)
	dbSiakadPort := getEnv("DB_SIAKAD_PORT", dbPort)
	dbSiakadUser := getEnv("DB_SIAKAD_USER", dbUser)
	dbSiakadPass := getEnv("DB_SIAKAD_PASS", dbPass)
	dbSiakadName := getEnv("DB_SIAKAD_NAME", dbName)
	dbSiakadSSL := getEnv("DB_SIAKAD_SSLMODE", dbSSL)

	return &Config{
		Port:            port,
		DBHost:          dbHost,
		DBPort:          dbPort,
		DBUser:          dbUser,
		DBPass:          dbPass,
		DBName:          dbName,
		DBSSLMode:       dbSSL,
		DBSiakadHost:    dbSiakadHost,
		DBSiakadPort:    dbSiakadPort,
		DBSiakadUser:    dbSiakadUser,
		DBSiakadPass:    dbSiakadPass,
		DBSiakadName:    dbSiakadName,
		DBSiakadSSLMode: dbSiakadSSL,
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
