package mySQL

import "os"

var DBDRIVER = "mysql"
var USERNAME = os.Getenv("MYSQL_ROOT_USERNAME")
var PASSWORD = os.Getenv("MYSQL_ROOT_PASSWORD")
var DBNAME = "super_nova_test_1"
var DBDATASOURCE = USERNAME + ":" + PASSWORD + "@/" + DBNAME
