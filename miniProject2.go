package main

import (
"fmt"
"gobot.io/x/gobot"
"gobot.io/x/gobot/drivers/aio"
"gobot.io/x/gobot/drivers/i2c"
g "gobot.io/x/gobot/platforms/dexter/gopigo3"
"gobot.io/x/gobot/platforms/raspi"
"time"
)

const (
	TOO_CLOSE  = 50
	TOO_FAR    = 150
	OUT_OF_RANGE = 200
)

func stop(gopigo3 *g.Driver) {
	err := gopigo3.SetMotorDps(g.MOTOR_LEFT+g.MOTOR_RIGHT, 0)
	if err != nil {
		fmt.Errorf("Error stopping the robot %+v", err)
	}
}

func forward(gopigo3 *g.Driver) {
	err := gopigo3.SetMotorDps(g.MOTOR_LEFT+g.MOTOR_RIGHT, 100)
	if err != nil {
		fmt.Errorf("Error moving forward %+v", err)
	}
}

func adjust_right(gopigo3 *g.Driver) {
	err := gopigo3.SetMotorDps(g.MOTOR_LEFT, 75)
	if err != nil {
		fmt.Errorf("Error turning left %+v", err)
	}
	err = gopigo3.SetMotorDps(g.MOTOR_RIGHT, 100)
	if err != nil {
		fmt.Errorf("Error turning left %+v", err)
	}
}

func adjust_left(gopigo3 *g.Driver) {
	err := gopigo3.SetMotorDps(g.MOTOR_LEFT, 100)
	if err != nil {
		fmt.Errorf("Error turning left %+v", err)
	}
	err = gopigo3.SetMotorDps(g.MOTOR_RIGHT, 75)
	if err != nil {
		fmt.Errorf("Error turning left %+v", err)
	}
}

func turn(gopigo3 *g.Driver) {
	err := gopigo3.SetMotorDps(g.MOTOR_LEFT, -180)
	if err != nil {
		fmt.Errorf("Error turning left %+v", err)
	}
	err = gopigo3.SetMotorDps(g.MOTOR_RIGHT, 180)
	if err != nil {
		fmt.Errorf("Error turning left %+v", err)
	}
}

func robotMainLoop(piProcessor *raspi.Adaptor, gopigo3 *g.Driver, lidarSensor *i2c.LIDARLiteDriver,

) {
	turned := false
	err := lidarSensor.Start()
	if err != nil {
		fmt.Println("error starting lidarSensor")
	}
	var secondCount = 0
	for { //loop forever
		lidarReading, err := lidarSensor.Distance()
		if err != nil {
			fmt.Println("Error reading lidar sensor %+v", err)
		}
		message := fmt.Sprintf("Lidar Reading: %d", lidarReading)

		fmt.Println(lidarReading)
		fmt.Println(message)
		time.Sleep(time.Second)
		//if robot isn't in range of the box move forward without counting
		if lidarReading > TOO_FAR && secondCount == 0{
			forward(gopigo3)
			time.Sleep(time.Second)
		//if robot is in the correct range of the box start counting for calculation
		}else if TOO_CLOSE < lidarReading && lidarReading < TOO_FAR {
			forward(gopigo3)
			time.Sleep(time.Second)
			secondCount += 1
		//if robot is too close to the box adjust slightly away from the box
		}else if lidarReading < TOO_CLOSE{
			adjust_right(gopigo3)
			time.Sleep(time.Second)
			secondCount += 1
		//if robot is too far from the box but still in range of the box slightly adjust towards the box
		}else if lidarReading > TOO_FAR && lidarReading < OUT_OF_RANGE && secondCount > 0{
			adjust_left(gopigo3)
			time.Sleep(time.Second)
			secondCount += 1
		//if robot passed the side of the box turn to measure next side
		}else if lidarReading > OUT_OF_RANGE && turned == false {
			err := gopigo3.SetMotorDps(g.MOTOR_LEFT+g.MOTOR_RIGHT, 200)
			if err != nil {
				fmt.Errorf("Error moving forward %+v", err)
			}
			turn(gopigo3)
			time.Sleep(time.Second)
			turned = true
		//if robot passes the edge of a box and has already turned meaning this is the second side then break
		}else if lidarReading > OUT_OF_RANGE && turned == true {
			break
		}
	}
	stop(gopigo3)
	var lengthOfBox float64 = float64(secondCount) * 100 * .5803
	fmt.Sprintf("The length of the box is: %d", lengthOfBox)

}

func main() {
	raspberryPi := raspi.NewAdaptor()
	gopigo3 := g.NewDriver(raspberryPi)
	lidarSensor := i2c.NewLIDARLiteDriver(raspberryPi)
	lightSensor := aio.NewGroveLightSensorDriver(gopigo3, "AD_2_1")
	workerThread := func() {
		robotMainLoop(raspberryPi, gopigo3, lidarSensor)
	}
	robot := gobot.NewRobot("Gopigo Pi4 Bot",
		[]gobot.Connection{raspberryPi},
		[]gobot.Device{gopigo3, lidarSensor, lightSensor},
		workerThread,
	)

	err := robot.Start()

	if err != nil {
		fmt.Errorf("Error starting Robot #{err}")
	}

}