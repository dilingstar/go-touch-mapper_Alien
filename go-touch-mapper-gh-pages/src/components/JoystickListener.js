import { useEffect, useRef } from "react";

export default function JoystickListener(props) {
    const indexButton = ["A", "B", "X", "Y", "LB", "RB", "LT", "RT", "SELECT", "START", "LS", "RS", "DPAD_UP", "DPAD_DOWN", "DPAD_LEFT", "DPAD_RIGHT", "HOME", "17", "18", "19", "20"];
    
    const connectedGamepad = useRef([]);
    const gplastStates = useRef({}); 

    const gamepadconnected = (e) => {
        console.log("gamepad connected", e.gamepad.index);
        connectedGamepad.current.push(e.gamepad.index);
        gplastStates.current[e.gamepad.index] = {
            buttons: e.gamepad.buttons.map(btn => false),
        };
    };

    const gamepaddisconnected = (e) => { 
        console.log("gamepad disconnected", e.gamepad.index);
        connectedGamepad.current = connectedGamepad.current.filter(x => x !== e.gamepad.index);
        delete gplastStates.current[e.gamepad.index];
    };

    const handelEvent = (btnIndex, downing) => { 
        const name = "BTN_" + indexButton[btnIndex];
        // [V3.2.0] 关键修改：传递第二个参数 isDown
        // 这样 ConfigManager 就能区分是按下还是抬起，从而正确维护按键池
        if (props.setDowningBtn) {
            props.setDowningBtn(name, downing);
        }
    };

    const stateChecker = () => { 
        const gamepads = navigator.getGamepads ? navigator.getGamepads() : [];
        
        for (let gpIndex of connectedGamepad.current) { 
            const gp = gamepads[gpIndex];
            if (!gp) continue;

            for (let i = 0; i < gp.buttons.length; i++) { 
                const pressed = gp.buttons[i].pressed;
                // 只有状态改变时才触发
                if (pressed !== gplastStates.current[gpIndex].buttons[i]) {
                    gplastStates.current[gpIndex].buttons[i] = pressed;
                    handelEvent(i, pressed);
                }
            }
        }
    };

    useEffect(() => {
        window.addEventListener("gamepadconnected", gamepadconnected);
        window.addEventListener("gamepaddisconnected", gamepaddisconnected);
        const interval = setInterval(stateChecker, 4);
        return () => {
            window.removeEventListener("gamepadconnected", gamepadconnected);
            window.removeEventListener("gamepaddisconnected", gamepaddisconnected);
            clearInterval(interval);
        };
    }, []);

    return <div style={{display:"none"}}/>;
}