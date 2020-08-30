package utils

import "fmt"

func makeFileName(dbName, suffix string, number uint64) string {
	return fmt.Sprintf("%s_%06d.%s", dbName, number, suffix)
}

func TableFileName(dbName string, number uint64) string {
	return makeFileName(dbName, "ldb", number)
}

func DescriptorFileName(dbName string, number uint64) string {
	return fmt.Sprintf("%s_MANIFEST-%06d", dbName, number)
}

func CurrentFileName(dbName string) string {
	return dbName + "_CURRENT"
}

func TempFileName(dbName string, number uint64) string {
	return makeFileName(dbName, "dbtmp", number)
}
