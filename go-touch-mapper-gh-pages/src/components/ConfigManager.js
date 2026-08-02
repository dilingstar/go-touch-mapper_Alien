import { useEffect, useRef, useState, useCallback } from "react";
import DraggableContainer from "./DraggableContainer";
import JoystickListener from "./JoystickListener";
import * as keyNameMapRaw from "./keynamemap.json";
import { produce } from "immer";
import {
    FixedIcon,
    GroupFixedIcon,
    WheelShow,
    ViewShow,
    ViewResetRadiusShow,
    ScrollSliderShow,
    VMouseResetPosShow,
} from "./UIcomponents";

import SettingsPanel from "./SettingsPanel"; 
import { safeCheckConfig, imageUrlToBase64 } from "./ConfigUtils";

// Version: V3.4.6

const keyNameMap = keyNameMapRaw.default || keyNameMapRaw;

export default function ConfigManager() {
    const [config, setConfig] = useState({
        "SCREEN": { "SIZE": [3200, 1440] },
        "MOUSE": { "SWITCH_KEYS": ["KEY_GRAVE"], "POINTER_SWITCH_KEYS": [], "POS": [0.52, 0.5], "SPEED": [0.3, 0.3] },
        "V_MOUSE_SETTINGS": { "ENABLE_INVERT_SCROLL": true, "RESET_POS": [0.5, 0.5], "MOUSE_SPEED": [1.0, 1.0], "LEFT_CLICK_KEYS": [], "RIGHT_CLICK_KEYS": [], "SCROLL_CONFIG": {} },
        "GLOBAL_RECOIL": { "ENABLE": false, "TRIGGER_KEYS": [], "SCOPE_KEYS": [], "RESET_SPEED_KEYS": [] },
        "WHEEL": { "POS": [0.17, 0.73], "WASD": ["KEY_W", "KEY_A", "KEY_S", "KEY_D"], "WHEEL_PLANET": { "PLANET_DYNAMIC_SPEED": {} }, "PLANET_CURVE": {} },
        "SCROLL_SLIDER": { "POS": [0.9, 0.5] },
        "KEY_JITTER": { "ENABLE": false, "AMOUNT": 0.013 },
        "KEY_MAPS": {},
        "MACROS": [], 
        "IMG": "data:image/webp;base64,"
    });

    const [tempImg, setTempImg] = useState(null);
    const [pluginConfig, setPluginConfig] = useState({});
    const [pluginValue, setPluginValue] = useState({});
    const [exportButtonText, setExportButtonText] = useState("更新配置");
    const [isExporting, setIsExporting] = useState(false);
    
    const activeKeysRef = useRef(new Set()); 
    const [virtualMouseKey, setVirtualMouseKey] = useState(null); 
    const [currentActiveKey, setCurrentActiveKey] = useState(null); 
    const [selectKEY, setSelectKEY] = useState(null); 
    
    const [visualSwitches, setVisualSwitches] = useState({
        keys: true,
        combos: true,
        viewCenter: true,
        viewRadius: true,
        vMouse: true,
        wheel: true,
        scroll: true,
        macros: true
    });

    const [imgSize, setImgSize] = useState([1, 1]);

    const getPostionValueX = useCallback((value) => { return parseInt(value * imgSize[0]) }, [imgSize]);
    const getPostionValueY = useCallback((value) => { return parseInt(value * imgSize[1]) }, [imgSize]);
    
    const viewCenterSetting = useRef(false);
    const scrollSliderPosSelecting = useRef(false);
    const vMouseResetPosSetting = useRef(false);
    const wheelPosSelecting = useRef(false);
    
    const [addingMacroPoint, setAddingMacroPoint] = useState(null); 
    const [addingPointKey, setAddingPointKey] = useState(null);
    const [settingPosBKey, setSettingPosBKey] = useState(null);
    const [settingSmartToggleIndexKey, setSettingSmartToggleIndexKey] = useState(null); 
    
    const updateCurrentActiveKey = useCallback(() => {
        const physicalKeys = Array.from(activeKeysRef.current);
        let combined = [...physicalKeys];
        
        if (virtualMouseKey) {
            if (!combined.includes(virtualMouseKey)) {
                combined.push(virtualMouseKey);
            }
        }

        if (combined.length > 0) {
            const displayKey = combined.join("+");
            setCurrentActiveKey(displayKey);
            if (activeKeysRef.current.size > 0 || virtualMouseKey) {
                setSelectKEY(displayKey);
            }
        } else {
            setCurrentActiveKey(null);
            setSelectKEY(null); 
        }
    }, [virtualMouseKey]);

    useEffect(() => {
        updateCurrentActiveKey();
    }, [virtualMouseKey, updateCurrentActiveKey]);

    const handleInputKey = useCallback((codeKey, isDown) => {
        if (!codeKey) return;
        if (isDown) {
            activeKeysRef.current.add(codeKey);
        } else {
            activeKeysRef.current.delete(codeKey);
        }
        updateCurrentActiveKey();
    }, [updateCurrentActiveKey]);

    const handleMouseBtnClick = (btn) => {
        setVirtualMouseKey(btn); 
    };

    const handleFileChange = (e) => {
        const reads = new FileReader();
        reads.readAsDataURL(document.getElementById('fileInput').files[0]);
        reads.onload = async function (e) {
            const bas64STR = await imageUrlToBase64(this.result);
            setTempImg(null);
            setConfig(produce(draft => { draft.IMG = bas64STR }));
        };
    };

    const getRemoteApiImg = async (url) => {
        const bas64STR = await imageUrlToBase64(url);
        setTempImg(null);
        setConfig(produce(draft => { draft.IMG = bas64STR }));
    };

    const imgLoaded = () => {
        setImgSize([document.getElementById("img").width, document.getElementById("img").height]);
        setConfig(produce(draft => { draft.SCREEN.SIZE = [document.getElementById("img").naturalWidth, document.getElementById("img").naturalHeight] }));
    };

    const exportJSON = () => {
        setExportButtonText("配置更新中...");
        setIsExporting(true);
        fetch('/configure/set', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                "config": config,
                "plugin": pluginValue
            })
        }).then(resp => resp.text()).then(text => {
            setExportButtonText(text);
            setTimeout(() => { setExportButtonText("更新配置"); setIsExporting(false); }, 1000);
        }).catch(err => {
            setExportButtonText(String(err));
            setTimeout(() => { setExportButtonText("更新配置"); setIsExporting(false); }, 1000);
        });
    };

    const handelImgClick = (e) => {
        const rect = document.getElementById("img").getBoundingClientRect();
        const key = selectKEY; 
        const x = (e.clientX - rect.left) / document.getElementById("img").width;
        const y = (e.clientY - rect.top) / document.getElementById("img").height;
        
        if (x > 1 || y > 1) return;

        if (addingMacroPoint) {
            setConfig(produce(draft => {
                if (!draft.MACROS) return;
                const macro = draft.MACROS.find(m => m.ID === addingMacroPoint.macroId);
                if (macro) {
                    let targetList;
                    if (addingMacroPoint.eventType === "PRESS") targetList = macro.PRESS_EVENTS;
                    else targetList = macro.RELEASE_EVENTS;
                    
                    if (targetList && targetList[addingMacroPoint.eventIndex]) {
                        const evt = targetList[addingMacroPoint.eventIndex];
                        if (addingMacroPoint.isSwipe) {
                            if (!evt.POS_LIST) evt.POS_LIST = [];
                            evt.POS_LIST.push([x, y]);
                        } else {
                            evt.POS = [x, y];
                        }
                    }
                }
            }));
            setAddingMacroPoint(null);
            return;
        }

        if (addingPointKey) {
            setConfig(produce(draft => {
                if(draft.KEY_MAPS[addingPointKey]) {
                    if(!draft.KEY_MAPS[addingPointKey].POS_S) draft.KEY_MAPS[addingPointKey].POS_S = [];
                    draft.KEY_MAPS[addingPointKey].POS_S.push([x, y]);
                }
            }));
            setAddingPointKey(null);
            return;
        }
        if (settingPosBKey) {
            setConfig(produce(draft => {
                if(draft.KEY_MAPS[settingPosBKey]) draft.KEY_MAPS[settingPosBKey].POS_B = [x, y];
            }));
            setSettingPosBKey(null);
            return;
        }
        
        if (settingSmartToggleIndexKey) {
            setConfig(produce(draft => {
                const mapObj = draft.KEY_MAPS[settingSmartToggleIndexKey.key];
                if(mapObj && mapObj.POS_S) {
                    mapObj.POS_S[settingSmartToggleIndexKey.index] = [x, y];
                }
            }));
            setSettingSmartToggleIndexKey(null);
            return;
        }

        if (viewCenterSetting.current) { setConfig(produce(draft => { draft.MOUSE.POS = [x, y] })); viewCenterSetting.current = false; return; }
        if (scrollSliderPosSelecting.current) { setConfig(produce(draft => { draft.SCROLL_SLIDER.POS = [x, y] })); scrollSliderPosSelecting.current = false; return; }
        if (vMouseResetPosSetting.current) { setConfig(produce(draft => { draft.V_MOUSE_SETTINGS.RESET_POS = [x, y] })); vMouseResetPosSetting.current = false; return; }
        if (wheelPosSelecting.current) { setConfig(produce(draft => { draft.WHEEL.POS = [x, y] })); wheelPosSelecting.current = false; return; }

        if (key !== null) {
            setConfig(produce(draft => {
                if (draft.KEY_MAPS[key]) {
                    // [V3.4.6 核心修复] 判断键类型，如果是 SMART_TOGGLE，则修改它的 A点 (POS_S[0])
                    if (draft.KEY_MAPS[key].TYPE === "SMART_TOGGLE") {
                        if (!draft.KEY_MAPS[key].POS_S || draft.KEY_MAPS[key].POS_S.length < 2) {
                            draft.KEY_MAPS[key].POS_S = [[x, y], [x + 0.05, y + 0.05]];
                        } else {
                            draft.KEY_MAPS[key].POS_S[0] = [x, y];
                        }
                    } else {
                        // 其他普通的基于 POS 的类型
                        draft.KEY_MAPS[key].POS = [x, y];
                    }
                } else {
                    if (key === "REL_WHEEL_UP" || key === "REL_WHEEL_DOWN") {
                        draft.KEY_MAPS[key] = { "TYPE": "CLICK", "POS": [x, y], "INTERVAL": [18] };
                    } else {
                        draft.KEY_MAPS[key] = { "TYPE": "PRESS", "POS": [x, y] };
                    }
                }
            }));
            
            if (virtualMouseKey) {
                setVirtualMouseKey(null);
            }
            setSelectKEY(null);
        }
    };

    useEffect(() => {
        fetch("/configure/get")
            .then(resp => resp.json())
            .then(data => {
                const safeData = safeCheckConfig(data);
                if (!safeData.MACROS) safeData.MACROS = [];
                setConfig(safeData);
            })
            .catch(console.error);

        fetch("/plugin/configure/getConfig")
            .then(resp => resp.json())
            .then(userConfig => {
                fetch("/plugin/configure/getTemplate")
                    .then(resp => resp.json())
                    .then(pluginTemplate => {
                        setPluginValue(userConfig);
                        setPluginConfig(pluginTemplate);
                    });
            })
            .catch(console.error);
    }, []);

    useEffect(() => {
        document.onkeydown = (e) => {
            if (e.repeat === false && window.stopPreventDefault !== true) {
                const codeKey = keyNameMap[e.code.toLowerCase()];
                if (codeKey) handleInputKey(codeKey, true);
            }
        };

        document.onkeyup = (e) => {
            if (window.stopPreventDefault !== true) {
                const codeKey = keyNameMap[e.code.toLowerCase()];
                if (codeKey) handleInputKey(codeKey, false);
            }
        };

        document.oncontextmenu = (e) => e.preventDefault();

        const handleResize = () => {
            if (document.getElementById("img")) {
                setImgSize([document.getElementById("img").width, document.getElementById("img").height]);
            }
        };
        window.addEventListener("resize", handleResize);

        return () => {
            document.onkeydown = null;
            document.onkeyup = null;
            document.oncontextmenu = null;
            window.removeEventListener("resize", handleResize);
        };
    }, [handleInputKey]);

    const KeyShow = ({ data }) => {
        const isCombo = data.KEY && data.KEY.includes("+");
        
        if (isCombo && !visualSwitches.combos) return null;
        if (!isCombo && !visualSwitches.keys) return null;

        const posBasedTypes = ["PRESS", "AUTO_FIRE", "CLICK", "SYNC_VIEW_RESET", "CLICK_VIEW_RESET", "BACKPACK_TOGGLE", "SEQUENTIAL_PRESS", "SYNC_BACKPACK", "PRESS_RELEASE_CLICK"];
        const showPos = posBasedTypes.includes(data["TYPE"]);
        const multiPosTypes = ["MULT_PRESS", "DRAG", "SEQUENTIAL_PRESS", "BACKPACK_TOGGLE", "SYNC_BACKPACK", "PRESS_RELEASE_CLICK", "SMART_TOGGLE"];
        const showMultiPos = multiPosTypes.includes(data["TYPE"]);

        let points = [];
        if (showPos && !showMultiPos) points = [data.POS];
        else if (["MULT_PRESS", "DRAG", "SMART_TOGGLE"].includes(data.TYPE)) points = data.POS_S || [];
        else if (data.TYPE === "SEQUENTIAL_PRESS") points = [data.POS, ...(data.POS_S || [])];
        else if (["BACKPACK_TOGGLE", "SYNC_BACKPACK", "PRESS_RELEASE_CLICK"].includes(data.TYPE)) points = [data.POS, (data.POS_B || data.POS)];

        const safePoints = points.filter(p => p && p.length >= 2);
        
        let randomRadius = undefined;
        if (config.KEY_JITTER.ENABLE) randomRadius = getPostionValueX(config.KEY_JITTER.AMOUNT);
        if (data.TYPE === "SYNC_VIEW_RESET" || data.TYPE === "CLICK_VIEW_RESET") {
            if (config.MOUSE.VIEW_RANDOM_RESET_ENABLE) {
                randomRadius = getPostionValueX(config.MOUSE.VIEW_RANDOM_RESET_RADIUS * 0.5);
            }
        }

        return <div>
            {showPos && !showMultiPos && data.POS && data.POS.length >= 2 ? 
                <FixedIcon 
                    x={getPostionValueX(data["POS"][0])} 
                    y={getPostionValueY(data["POS"][1])} 
                    text={data["KEY"]} 
                    type={data["TYPE"]}
                    random_radius={randomRadius}
                /> 
                : null
            }
            {showMultiPos ? 
                <GroupFixedIcon 
                    pos_s={safePoints.map(([x, y]) => [getPostionValueX(x), getPostionValueY(y)])} 
                    text={data["KEY"]} 
                    type={data["TYPE"]}
                    random_radius={randomRadius}
                /> 
                : null
            }
        </div>
    };

    const MacroShow = ({ macro }) => {
        if (!visualSwitches.macros) return null;
        let randomRadius = undefined;
        if (config.KEY_JITTER.ENABLE) randomRadius = getPostionValueX(config.KEY_JITTER.AMOUNT);

        const triggerLabel = (macro.TRIGGER_KEY || "").replace(/KEY_/g, "").replace(/BTN_/g, "");
        
        const renderEvents = (events, prefixSymbol) => {
            if (!events) return null;
            return events.map((evt, idx) => {
                if (evt.TYPE === "CLICK" && evt.POS) {
                    return <FixedIcon 
                        key={`${macro.ID}_${prefixSymbol}_${idx}`}
                        x={getPostionValueX(evt.POS[0])}
                        y={getPostionValueY(evt.POS[1])}
                        text={`${triggerLabel}\n${prefixSymbol}_${idx+1}`}
                        type="MACRO_CLICK"
                        random_radius={randomRadius}
                    />
                } else if (evt.TYPE === "SWIPE" && evt.POS_LIST && evt.POS_LIST.length >= 2) {
                    return <GroupFixedIcon 
                        key={`${macro.ID}_${prefixSymbol}_${idx}`}
                        pos_s={evt.POS_LIST.map(([x, y]) => [getPostionValueX(x), getPostionValueY(y)])}
                        text={`${triggerLabel}\n${prefixSymbol}_${idx+1}`}
                        type="MACRO_SWIPE"
                        random_radius={randomRadius}
                    />
                }
                return null;
            });
        };

        let pressSymbol = "↓";
        let releaseSymbol = "↑";
        
        if (macro.TYPE === "HOLD_LOOP") {
            pressSymbol = "↓↑";
        } else if (macro.TYPE === "TOGGLE_LOOP") {
            pressSymbol = "↺";
        }

        return (
            <div>
                {macro.PRESS_EVENTS && renderEvents(macro.PRESS_EVENTS, pressSymbol)}
                {macro.RELEASE_EVENTS && renderEvents(macro.RELEASE_EVENTS, releaseSymbol)}
            </div>
        );
    };

    return (
        <div style={{ width: '100vw', height: '100vh', backgroundColor: '#00796B' }}>
            <JoystickListener setDowningBtn={handleInputKey} />
            <input id="fileInput" type="file" style={{ display: "none" }} accept="image/*" onChange={handleFileChange} />
            <img id="img" src={tempImg || config["IMG"]} style={{ width: "100vw", left: 0, top: 0 }} onClick={handelImgClick} onLoad={imgLoaded} alt="mapping-bg" />
            
            <DraggableContainer>
                <SettingsPanel 
                    config={config} setConfig={setConfig}
                    pluginConfig={pluginConfig} pluginValue={pluginValue} setPluginValue={setPluginValue}
                    exportJSON={exportJSON} exportButtonText={exportButtonText} isExporting={isExporting}
                    selectKEY={selectKEY} setSelectKEY={setSelectKEY} 
                    
                    setViewCenterSetting={() => { viewCenterSetting.current = true; }} isViewCenterSetting={() => viewCenterSetting.current}
                    setScrollSliderPosSelecting={() => { scrollSliderPosSelecting.current = true; }} isScrollSliderPosSelecting={() => scrollSliderPosSelecting.current}
                    setVMouseResetPosSetting={() => { vMouseResetPosSetting.current = true; }} isVMouseResetPosSetting={() => vMouseResetPosSetting.current}
                    setWheelPosSelecting={() => { wheelPosSelecting.current = true; }} isWheelPosSelecting={() => wheelPosSelecting.current}
                    
                    onStartAddingPoint={setAddingPointKey} isAddingPoint={(key) => addingPointKey === key}
                    onStartSettingPosB={setSettingPosBKey} isSettingPosB={(key) => settingPosBKey === key}
                    onStartSettingSmartToggleIndex={setSettingSmartToggleIndexKey} isSettingSmartToggleIndex={(key, idx) => settingSmartToggleIndexKey && settingSmartToggleIndexKey.key === key && settingSmartToggleIndexKey.index === idx}
                    
                    getRemoteApiImg={getRemoteApiImg} openFileInput={() => { document.getElementById('fileInput').click(); }}
                    
                    visualSwitches={visualSwitches} setVisualSwitches={setVisualSwitches}
                    setTempImg={setTempImg}

                    currentActiveKey={currentActiveKey}
                    addingMacroPoint={addingMacroPoint} setAddingMacroPoint={setAddingMacroPoint}
                    
                    onMouseBtnClick={handleMouseBtnClick}
                />
            </DraggableContainer>

            {Object.keys(config["KEY_MAPS"]).map((keycode) => 
                <KeyShow key={keycode} data={{ ...config["KEY_MAPS"][keycode], "KEY": keycode }} />
            )}

            {config.MACROS && config.MACROS.map((macro) => 
                <MacroShow key={macro.ID} macro={macro} />
            )}
            
            {visualSwitches.wheel && <WheelShow
                x={getPostionValueX(config["WHEEL"]["POS"][0])}
                y={getPostionValueY(config["WHEEL"]["POS"][1])}
                range={getPostionValueX(config["WHEEL"]["RANGE"])}
                shift_range={config["WHEEL"]["SHIFT_RANGE_ENABLE"] ? getPostionValueX(config["WHEEL"]["SHIFT_RANGE"]) : 0}
                bezier_enable={config["WHEEL"]["BEZIER_ENABLE"]}
                random_point_enable={config["WHEEL"]["RANDOM_POINT_ENABLE"]}
                random_start_radius={getPostionValueX(config["WHEEL"]["RANDOM_START_RADIUS"])}
                random_end_radius={getPostionValueX(config["WHEEL"]["RANDOM_END_RADIUS"])}
                random_shift_end_radius={getPostionValueX(config["WHEEL"]["RANDOM_SHIFT_END_RADIUS"])}
            />}
            
            {visualSwitches.viewCenter && <ViewShow x={getPostionValueX(config["MOUSE"]["POS"][0])} y={getPostionValueY(config["MOUSE"]["POS"][1])} />}
            
            {visualSwitches.viewRadius && <ViewResetRadiusShow
                 x={getPostionValueX(config["MOUSE"]["POS"][0])}
                 y={getPostionValueY(config["MOUSE"]["POS"][1])}
                 radius={getPostionValueX(config.MOUSE.VIEW_RESET_RADIUS)} 
                 enable={config.MOUSE.VIEW_RESET_RADIUS_ENABLE}
                 random_enable={config.MOUSE.VIEW_RANDOM_RESET_ENABLE}
                 random_radius={getPostionValueX(config.MOUSE.VIEW_RANDOM_RESET_RADIUS)}
            />}
            
            {visualSwitches.scroll && <ScrollSliderShow
                x={getPostionValueX(config.SCROLL_SLIDER.POS[0])}
                y={getPostionValueY(config.SCROLL_SLIDER.POS[1])}
                lengthUp={getPostionValueY(config.SCROLL_SLIDER.LENGTH_UP)}
                lengthDown={getPostionValueY(config.SCROLL_SLIDER.LENGTH_DOWN)}
                enable={config.SCROLL_SLIDER.ENABLE}
                random_enable={config.SCROLL_SLIDER.RANDOM_START_ENABLE}
                random_start_radius={getPostionValueX(config.SCROLL_SLIDER.RANDOM_START_RADIUS)}
            />}
            
            {visualSwitches.vMouse && <VMouseResetPosShow
                x={getPostionValueX(config.V_MOUSE_SETTINGS.RESET_POS[0])}
                y={getPostionValueY(config.V_MOUSE_SETTINGS.RESET_POS[1])}
                enable={true}
            />}
        </div>
    );
}