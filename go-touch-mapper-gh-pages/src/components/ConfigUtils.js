import { produce } from "immer";

// Version: V3.4.5

export function imageUrlToBase64(url) {
    return new Promise((resolve, reject) => {
        const img = new Image();
        img.crossOrigin = "Anonymous";
        img.src = url;
        img.onload = () => {
            const canvas = document.createElement('canvas');
            const ctx = canvas.getContext('2d');
            canvas.width = img.width;
            canvas.height = img.height;
            ctx.drawImage(img, 0, 0);
            try {
                const base64String = canvas.toDataURL('image/png');
                resolve(base64String);
            } catch (e) {
                reject(`转换失败: ${e}`);
            }
        };
        img.onerror = (err) => {
            reject(`图片加载失败: ${err}`);
        };
    });
}

export const safeCheckConfig = (data) => {
    return produce(data, draft => {
        // 1. Mouse & Switch Keys
        if (!draft.MOUSE) draft.MOUSE = {};
        if (!draft.MOUSE.SWITCH_KEYS) draft.MOUSE.SWITCH_KEYS = ["KEY_GRAVE"];
        if (!draft.MOUSE.POINTER_SWITCH_KEYS) draft.MOUSE.POINTER_SWITCH_KEYS = [];
        
        if (draft.MOUSE.VIEW_AUTO_RELEASE_ENABLE === undefined) {
            if (draft.MOUSE.VIEW_AUTO_RELEASE_MS && draft.MOUSE.VIEW_AUTO_RELEASE_MS > 0) {
                draft.MOUSE.VIEW_AUTO_RELEASE_ENABLE = true;
            } else {
                draft.MOUSE.VIEW_AUTO_RELEASE_ENABLE = false;
                draft.MOUSE.VIEW_AUTO_RELEASE_MS = 200;
            }
        }
        if (draft.MOUSE.VIEW_RESET_RADIUS_ENABLE === undefined) draft.MOUSE.VIEW_RESET_RADIUS_ENABLE = false;
        if (draft.MOUSE.VIEW_RESET_RADIUS === undefined) draft.MOUSE.VIEW_RESET_RADIUS = 0.1;
        if (draft.MOUSE.VIEW_RESET_RADIUS_THICKNESS === undefined) draft.MOUSE.VIEW_RESET_RADIUS_THICKNESS = 0.005;
        if (draft.MOUSE.VIEW_RANDOM_RESET_ENABLE === undefined) draft.MOUSE.VIEW_RANDOM_RESET_ENABLE = false;
        if (draft.MOUSE.VIEW_RANDOM_RESET_RADIUS === undefined) draft.MOUSE.VIEW_RANDOM_RESET_RADIUS = 0.01;
        if (draft.MOUSE.VIEW_DELAY_RESET_ENABLE === undefined) draft.MOUSE.VIEW_DELAY_RESET_ENABLE = false;
        if (draft.MOUSE.VIEW_DELAY_RESET_MS === undefined) draft.MOUSE.VIEW_DELAY_RESET_MS = 20;
        if (draft.MOUSE.VIEW_DELAY_RESET_RANDOM_ENABLE === undefined) draft.MOUSE.VIEW_DELAY_RESET_RANDOM_ENABLE = false;
        if (draft.MOUSE.VIEW_DELAY_RESET_MIN_MS === undefined) draft.MOUSE.VIEW_DELAY_RESET_MIN_MS = 10;

        // 2. V_MOUSE_SETTINGS
        const defaultVMouse = {
            "ENABLE_INVERT_SCROLL": true, "RESET_POS": [0.5, 0.5], "MOUSE_SPEED": [1.0, 1.0],
            "LEFT_CLICK_KEYS": [], "RIGHT_CLICK_KEYS": [], // [V3.4.5] 新增手柄等价点击键
            "SCROLL_CONFIG": {
                "RELEASE_DELAY_MS": 50, "NON_RESET_MS": 300, "RESET_DELAY_MS": 50, "SPEED": 1.0,
                "CURVE_ENABLE": false, "CURVE_AMOUNT": 0.005, "CURVE_FREQ": 1.0, 
                "DYNAMIC_NOISE_ENABLE": false, "DYNAMIC_NOISE_AMOUNT": 0.002
            }
        };
        if (!draft.V_MOUSE_SETTINGS) draft.V_MOUSE_SETTINGS = defaultVMouse;
        if (!draft.V_MOUSE_SETTINGS.RESET_POS) draft.V_MOUSE_SETTINGS.RESET_POS = [0.5, 0.5];
        if (!draft.V_MOUSE_SETTINGS.MOUSE_SPEED) draft.V_MOUSE_SETTINGS.MOUSE_SPEED = [1.0, 1.0];
        if (!draft.V_MOUSE_SETTINGS.LEFT_CLICK_KEYS) draft.V_MOUSE_SETTINGS.LEFT_CLICK_KEYS = [];
        if (!draft.V_MOUSE_SETTINGS.RIGHT_CLICK_KEYS) draft.V_MOUSE_SETTINGS.RIGHT_CLICK_KEYS = [];
        if (!draft.V_MOUSE_SETTINGS.SCROLL_CONFIG) draft.V_MOUSE_SETTINGS.SCROLL_CONFIG = defaultVMouse.SCROLL_CONFIG;
        
        // 3. GLOBAL_RECOIL
        const defaultRecoil = {
            "ENABLE": false, "TRIGGER_KEYS": ["BTN_LEFT"], "SCOPE_MODE": false, "SCOPE_KEYS": ["BTN_RIGHT"],
            "BASE_SPEED": 0, "RESET_SPEED_KEYS": ["BTN_MIDDLE"]
        };
        if (!draft.GLOBAL_RECOIL) draft.GLOBAL_RECOIL = defaultRecoil;
        if (!draft.GLOBAL_RECOIL.TRIGGER_KEYS) draft.GLOBAL_RECOIL.TRIGGER_KEYS = [];
        if (!draft.GLOBAL_RECOIL.SCOPE_KEYS) draft.GLOBAL_RECOIL.SCOPE_KEYS = [];
        if (!draft.GLOBAL_RECOIL.RESET_SPEED_KEYS) draft.GLOBAL_RECOIL.RESET_SPEED_KEYS = [];

        // 4. WHEEL
        if (!draft.WHEEL) draft.WHEEL = {};
        if (!draft.WHEEL.WHEEL_PLANET) draft.WHEEL.WHEEL_PLANET = { ENABLE: false, RADIUS: 0.015, SPEED: 1.5 };
        if (!draft.WHEEL.WHEEL_PLANET.PLANET_DYNAMIC_SPEED) draft.WHEEL.WHEEL_PLANET.PLANET_DYNAMIC_SPEED = { ENABLE: false, MIN_SPEED: 0.5, FREQUENCY: 1.0 };
        if (!draft.WHEEL.PLANET_CURVE) draft.WHEEL.PLANET_CURVE = { ENABLE: false, CURVE_AMOUNT: 0.005, CURVE_FREQUENCY: 1.0 };

        if (draft.WHEEL.DELAY_RESET_MS === undefined) draft.WHEEL.DELAY_RESET_MS = 50;
        if (draft.WHEEL.STEP_SPEED === undefined) draft.WHEEL.STEP_SPEED = 60;
        if (draft.WHEEL.BEZIER_ENABLE === undefined) draft.WHEEL.BEZIER_ENABLE = false;
        if (draft.WHEEL.BEZIER_SPEED === undefined) draft.WHEEL.BEZIER_SPEED = 60.0;
        if (draft.WHEEL.BEZIER_CURVE_AMOUNT === undefined) draft.WHEEL.BEZIER_CURVE_AMOUNT = 0.5;
        if (draft.WHEEL.BEZIER_DYNAMIC_CURVE === undefined) draft.WHEEL.BEZIER_DYNAMIC_CURVE = 0.0;
        if (draft.WHEEL.RANDOM_POINT_ENABLE === undefined) draft.WHEEL.RANDOM_POINT_ENABLE = false;
        if (draft.WHEEL.RANDOM_START_RADIUS === undefined) draft.WHEEL.RANDOM_START_RADIUS = 0.01;
        if (draft.WHEEL.RANDOM_END_RADIUS === undefined) draft.WHEEL.RANDOM_END_RADIUS = 0.01;
        if (draft.WHEEL.RANDOM_SHIFT_END_RADIUS === undefined) draft.WHEEL.RANDOM_SHIFT_END_RADIUS = 0.015;
        if (draft.WHEEL.EASING_ENABLE === undefined) draft.WHEEL.EASING_ENABLE = false;
        if (draft.WHEEL.EASING_IN === undefined) draft.WHEEL.EASING_IN = 0.2;
        if (draft.WHEEL.EASING_OUT === undefined) draft.WHEEL.EASING_OUT = 0.2;
        if (draft.WHEEL.NOISE_ENABLE === undefined) draft.WHEEL.NOISE_ENABLE = false;
        if (draft.WHEEL.NOISE_INTENSITY === undefined) draft.WHEEL.NOISE_INTENSITY = 0.002;
        if (draft.WHEEL.WHEEL_PLANET.PLANET_NOISE_INTENSITY === undefined) draft.WHEEL.WHEEL_PLANET.PLANET_NOISE_INTENSITY = 0.002;
        
        if (draft.WHEEL.SHIFT_PRESS_TOGGLE === undefined) draft.WHEEL.SHIFT_PRESS_TOGGLE = false;
        if (draft.WHEEL.SHIFT_RELEASE_TOGGLE === undefined) draft.WHEEL.SHIFT_RELEASE_TOGGLE = false;
        if (draft.WHEEL.TEMP_PENETRATION_ENABLE === undefined) draft.WHEEL.TEMP_PENETRATION_ENABLE = false;

        // 5. SCROLL_SLIDER
        const defaultScroll = {
            "ENABLE": false, "POS": [0.9, 0.5], "LENGTH_UP": 0.2, "LENGTH_DOWN": 0.2, "TIMEOUT_S": 3, "SPEED": 1.0,
            "RANDOM_START_ENABLE": false, "RANDOM_START_RADIUS": 0.005, 
            "CURVE_ENABLE": false, "CURVE_AMOUNT": 0.005, "CURVE_FREQUENCY": 1.0, 
            "DELAY_RESET_MS": 20, "DELAY_RANDOM_ENABLE": false, "DELAY_RESET_MIN_MS": 10,
            "RELEASE_DELAY_MS": 50, "NOISE_ENABLE": false, "NOISE_INTENSITY": 0.002
        };
        if (!draft.SCROLL_SLIDER) draft.SCROLL_SLIDER = defaultScroll;
        for (let key in defaultScroll) {
            if (draft.SCROLL_SLIDER[key] === undefined) draft.SCROLL_SLIDER[key] = defaultScroll[key];
        }

        // 6. KEY_JITTER
        if (!draft.KEY_JITTER) draft.KEY_JITTER = { ENABLE: true, AMOUNT: 0.003 };

        // 7. KEY_MAPS 数据清洗
        if (draft.KEY_MAPS) {
            Object.keys(draft.KEY_MAPS).forEach(key => {
                const map = draft.KEY_MAPS[key];
                if (map.TYPE === "AUTO_FIRE") {
                    if (!map.INTERVAL) map.INTERVAL = [18, 20, 10, 10];
                    else if (map.INTERVAL.length < 4) {
                        map.INTERVAL[2] = map.INTERVAL[2] || 10;
                        map.INTERVAL[3] = map.INTERVAL[3] || 10;
                    }
                }
            });
        }

        // 8. MACROS
        if (!draft.MACROS) draft.MACROS = [];
        
        if (!draft.END_EXIT) draft.END_EXIT = { "EXIT_ENABLE": false, "EXIT_KEYS": [] };
    });
};