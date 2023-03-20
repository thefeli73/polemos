package main

import (
	"fmt"
	"os"

	uuid "github.com/nu7hatch/gouuid"
	"github.com/thefeli73/polemos/mtd_aws"
	"github.com/thefeli73/polemos/state"
)
func main() {
	fmt.Println("Starting Polemos")

	config, err := state.Load_conf("config.yaml")
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	//aws_config := mtd_aws.New_config(config.AWS.Region)
    //mtd_aws.Instance_info(aws_config, config.AWS.InstanceID)
	//mtd_aws.Instances(config.AWS.Region)
	u, _ := uuid.NewV4() //blank is to send errors to the void
	_=u
	//fmt.Println(u)


	for _, region := range config.AWS.Regions {
        fmt.Println("Listing instances in region:", region)
		aws_config := mtd_aws.New_config(region, config.AWS.Credentials_path)
        instances, err := mtd_aws.Instances(aws_config)
        if err != nil {
            fmt.Println("Error listing instances:", err)
            continue
        }

        for _, instance := range instances {
            mtd_aws.Instance_info(aws_config, *instance.InstanceId)
        }
    }

}