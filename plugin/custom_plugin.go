//go:build ignore || plugin
// +build ignore plugin

package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

const (
	DEBUG_ENABLED = true
)

func DEBUG(msg string) {
	if DEBUG_ENABLED {
		fmt.Println(msg)
	}
}

func DEBUGF(format string, args ...interface{}) {
	if DEBUG_ENABLED {
		fmt.Printf(format+"\n", args...)
	}
}

var (
	CurveStrength   float64 = 0.1    // 曲线强度 (0-1之间，0为直线，1为强曲线)
	JitterIntensity float64 = 0.3    // 抖动强度
	JitterFrequency float64 = 8.0    // 抖动频率
	MinPoints       int     = 3      // 最少点数
	MaxPoints       int     = 20     // 最多点数
	PointsPerUnit   float64 = 1.0    // 每单位距离的点数
	EasingType      string  = "quad" // 缓动类型: "quad", "cubic", "sine"
)

type Point struct {
	X int32
	Y int32
}

// GenerateSwipeDisplacements 生成从起点到终点的位移量数组
// 返回从起点开始的每次位移到下一个点的坐标差值
func GenerateSwipeDisplacements(startX, startY, endX, endY int32) [][]int32 {
	// 将int32转换为float64进行计算
	sx, sy := float64(startX), float64(startY)
	ex, ey := float64(endX), float64(endY)
	dx := ex - sx
	dy := ey - sy
	distance := math.Sqrt(dx*dx + dy*dy)
	// 根据距离确定点数
	numPoints := int(distance / PointsPerUnit)
	if numPoints < MinPoints {
		numPoints = MinPoints
	}
	if numPoints > MaxPoints {
		numPoints = MaxPoints
	}
	// 生成绝对坐标点
	absolutePoints := generateAbsolutePoints(sx, sy, ex, ey, numPoints)
	return convertToPosList(absolutePoints, startX, startY)
}

func convertToPosList(absolutePoints []Point, startX, startY int32) [][]int32 {
	if len(absolutePoints) == 0 {
		return [][]int32{}
	}
	posList := make([][]int32, 0, len(absolutePoints))
	for i := range absolutePoints {
		posList = append(posList, []int32{
			absolutePoints[i].X,
			absolutePoints[i].Y,
		})
	}
	return posList
}

// generateAbsolutePoints 生成绝对坐标点
func generateAbsolutePoints(startX, startY, endX, endY float64, numPoints int) []Point {
	points := make([]Point, 0, numPoints)

	dx := endX - startX
	dy := endY - startY
	distance := math.Sqrt(dx*dx + dy*dy)

	// 初始化随机种子
	rand.Seed(time.Now().UnixNano())

	// 计算控制点 - 控制曲线的弯曲程度
	control1X := startX + dx*0.4 + rand.Float64()*distance*CurveStrength*2 - distance*CurveStrength
	control1Y := startY + dy*0.4 + rand.Float64()*distance*CurveStrength*2 - distance*CurveStrength
	control2X := startX + dx*0.6 + rand.Float64()*distance*CurveStrength*2 - distance*CurveStrength
	control2Y := startY + dy*0.6 + rand.Float64()*distance*CurveStrength*2 - distance*CurveStrength

	last_x := int32(0)
	last_y := int32(0)
	for i := 0; i < numPoints; i++ {
		t := float64(i) / float64(numPoints-1)
		easedT := applyEasing(t)
		x := cubicBezier(startX, control1X, control2X, endX, easedT)
		y := cubicBezier(startY, control1Y, control2Y, endY, easedT)
		if t > 0.15 && t < 0.85 {
			noise := math.Sin(t*JitterFrequency*math.Pi) * distance * JitterIntensity
			x += noise * (rand.Float64()*2 - 1)
			y += noise * (rand.Float64()*2 - 1)
		}
		rx := int32(math.Round(x))
		ry := int32(math.Round(y))
		if last_x != rx || last_y != ry {
			last_x = rx
			last_y = ry
			points = append(points, Point{
				X: rx,
				Y: ry,
			})
		}

	}

	return points
}

// cubicBezier 三次贝塞尔曲线计算
func cubicBezier(p0, p1, p2, p3, t float64) float64 {
	oneMinusT := 1 - t
	return oneMinusT*oneMinusT*oneMinusT*p0 +
		3*oneMinusT*oneMinusT*t*p1 +
		3*oneMinusT*t*t*p2 +
		t*t*t*p3
}

// applyEasing 应用缓动函数
func applyEasing(t float64) float64 {
	switch EasingType {
	case "quad":
		// 二次缓入缓出
		if t < 0.5 {
			return 2 * t * t
		}
		return -1 + (4-2*t)*t
	case "cubic":
		// 三次缓入缓出
		if t < 0.5 {
			return 4 * t * t * t
		}
		return 1 - math.Pow(-2*t+2, 3)/2
	case "sine":
		// 正弦缓入缓出
		return -(math.Cos(math.Pi*t) - 1) / 2
	default:
		// 默认线性
		return t
	}
}

const (
	PluginID             string = "lty.go-touch-mapper.default-plugin.v0.0.2"
	PluginConfigTemplate string = `{
        "CurveStrength": {
            "type": "int32",
            "default": 10,
            "min": 0,
            "max": 1000,
            "description": "曲线强度"
        },

		"JitterIntensity": {
            "type": "int32",
            "default": 30,
            "min": 0,
            "max": 1000,
            "description": "抖动强度"
        },

		"JitterFrequency": {
            "type": "int32",
            "default": 8,
            "min": 1,
            "max": 20,
            "description": "抖动频率"
        },
		"MinPoints": {
            "type": "int32",
            "default": 3,
            "min": 3,
            "max": 10,
            "description": "最少点数"
        },
		"MaxPoints": {
            "type": "int32",
            "default": 20,
            "min": 10,
            "max": 30,
            "description": "最多点数"
        },
		"PointsPerUnit": {
            "type": "int32",
            "default": 1,
            "min": 1,
            "max": 10,
            "description": "每单位距离的点数"
        },
        "EasingType": {
            "type": "select",
            "default": 0,
            "values": [
                "quad",
                "cubic",
                "sine"
            ],
            "description": "缓动类型: \"quad\", \"cubic\", \"sine\""
        }
    }`
)

func update_wheel_xy(last_x, last_y, target_x, target_y int32) (int32, int32) {
	const WHEEL_STEP_VAL = int32(120)
	if last_x == target_x && last_y == target_y {
		return last_x, last_y
	} else {
		x_rest := target_x - last_x
		y_rest := target_y - last_y
		total_rest := int32(math.Sqrt(float64(x_rest*x_rest + y_rest*y_rest)))
		if total_rest <= WHEEL_STEP_VAL {
			return target_x, target_y
		} else {
			return last_x + x_rest*WHEEL_STEP_VAL/total_rest, last_y + y_rest*WHEEL_STEP_VAL/total_rest
		}
	}
}

type CustomWheel struct {
	userConfig       map[string]interface{}
	wheelRadius      int32
	shiftWheelRadius int32
	centerX          int32
	center_y         int32
	screen_x         int32
	screen_y         int32
	//=========================================
	last_wheel_asix_x  int32
	last_wheel_asix_y  int32
	last_shift_pressed bool
	counter            int   //记录当xyshift不变的情况下调用的次数
	target_x           int32 //目标x坐标
	target_y           int32 //目标y坐标
	//=========================================
	posList [][]int32 //记录移动的路径

}

func initWheel() *CustomWheel {
	return &CustomWheel{}
}

func (cw *CustomWheel) update_user_config(userConfig map[string]interface{}) {
	cw.userConfig = userConfig
	CurveStrength = float64(cw.userConfig["CurveStrength"].(float64)) / 1000
	JitterIntensity = float64(cw.userConfig["JitterIntensity"].(float64)) / 1000
	JitterFrequency = float64(cw.userConfig["JitterFrequency"].(float64))
	MinPoints = int(cw.userConfig["MinPoints"].(float64))
	MaxPoints = int(cw.userConfig["MaxPoints"].(float64))
	PointsPerUnit = float64(cw.userConfig["PointsPerUnit"].(float64))
	EasingType = []string{"quad", "cubic", "sine"}[int(cw.userConfig["EasingType"].(float64))]
	DEBUGF("CurveStrength: %f, JitterIntensity: %f, JitterFrequency: %f, MinPoints: %d, MaxPoints: %d, PointsPerUnit: %f, EasingType: %s\n", CurveStrength, JitterIntensity, JitterFrequency, MinPoints, MaxPoints, PointsPerUnit, EasingType)

}
func (cw *CustomWheel) update_wheel_config(wheelRadius, shiftWheelRadius, centerX, center_y, screen_x, screen_y int32) {
	cw.wheelRadius = wheelRadius
	cw.shiftWheelRadius = shiftWheelRadius
	cw.centerX = centerX
	cw.center_y = center_y
	cw.screen_x = screen_x
	cw.screen_y = screen_y
}

func (cw *CustomWheel) get_wheel_move_target(wheel_pos_x, wheel_pos_y, wheel_axis_x, wheel_asix_y int32, shift_pressed bool) (int32, int32) {
	if cw.last_wheel_asix_x != wheel_axis_x || cw.last_wheel_asix_y != wheel_asix_y || cw.last_shift_pressed != shift_pressed {
		rand.Seed(time.Now().UnixNano())
		cw.counter = 0
		usingRadius := cw.wheelRadius
		if shift_pressed {
			usingRadius = cw.shiftWheelRadius
		}
		if wheel_axis_x*wheel_asix_y != 0 {
			usingRadius = usingRadius * 707 / 1000
		}
		cw.target_x = cw.centerX + wheel_axis_x*usingRadius + int32(rand.Float64()*CurveStrength*100)
		cw.target_y = cw.center_y + wheel_asix_y*usingRadius + int32(rand.Float64()*CurveStrength*100)
		if cw.target_x < 0 {
			cw.target_x = 0
		} else if cw.target_x > cw.screen_x {
			cw.target_x = cw.screen_x
		}
		if cw.target_y < 0 {
			cw.target_y = 0
		} else if cw.target_y > cw.screen_y {
			cw.target_y = cw.screen_y
		}
		DEBUGF("wheel_pos_x: %d, wheel_pos_y: %d, wheel_axis_x: %d, wheel_asix_y: %d, shift_pressed: %v, usingRadius: %d, target_x: %d, target_y: %d\n", wheel_pos_x, wheel_pos_y, wheel_axis_x, wheel_asix_y, shift_pressed, usingRadius, cw.target_x, cw.target_y)
		cw.posList = GenerateSwipeDisplacements(wheel_pos_x, wheel_pos_y, cw.target_x, cw.target_y)
		DEBUGF("posList: %v\n", cw.posList)
	}
	cw.last_wheel_asix_x = wheel_axis_x
	cw.last_wheel_asix_y = wheel_asix_y
	cw.last_shift_pressed = shift_pressed
	//====================================================================================
	if wheel_pos_x == cw.target_x && wheel_pos_y == cw.target_y {
		return wheel_pos_x, wheel_pos_y
	} else {
		if cw.counter >= len(cw.posList)-1 {
			return cw.target_x, cw.target_y
		} else {
			x, y := cw.posList[cw.counter][0], cw.posList[cw.counter][1]
			DEBUGF("counter: %d, x: %d, y: %d\n", cw.counter, x, y)
			cw.counter++
			return x, y
		}
	}
}
