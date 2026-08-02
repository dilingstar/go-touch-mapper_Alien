import React, { useState, useRef, useEffect } from 'react';
import { Button, Typography, Box, Paper } from "@mui/material";
import ViewInArIcon from '@mui/icons-material/ViewInAr';
import MouseIcon from '@mui/icons-material/Mouse';
import MapIcon from '@mui/icons-material/Map';
import RadioButtonCheckedIcon from '@mui/icons-material/RadioButtonChecked';
import HeightIcon from '@mui/icons-material/Height';
import GpsFixedIcon from '@mui/icons-material/GpsFixed';
import ExtensionIcon from '@mui/icons-material/Extension';
import SettingsIcon from '@mui/icons-material/Settings';
import KeyboardIcon from '@mui/icons-material/Keyboard';
import LinkIcon from '@mui/icons-material/Link';
import MovieFilterIcon from '@mui/icons-material/MovieFilter'; 

import GeneralSettings from "./settingspanel/general";
import { KeysSettings, ComboSettings } from "./settingspanel/keys";
import MapSettings from "./settingspanel/mapperswitch";
import ViewSettings from "./settingspanel/view";
import VMouseSettings from "./settingspanel/vmouse";
import WheelSettings from "./settingspanel/wheel";
import ScrollSettings from "./settingspanel/scroll";
import RecoilSettings from "./settingspanel/recoil";
import PluginSettings from "./settingspanel/plugin";
import MacroEditor from "./settingspanel/macroeditor";

// Version: V3.5.0

export default function SettingsPanel(props) {
    const { 
        config, setConfig, 
        pluginConfig, pluginValue, setPluginValue,
        exportJSON, exportButtonText, isExporting,
        selectKEY, setSelectKEY,
        setViewCenterSetting, isViewCenterSetting,
        setScrollSliderPosSelecting, isScrollSliderPosSelecting,
        setVMouseResetPosSetting, isVMouseResetPosSetting,
        setWheelPosSelecting, isWheelPosSelecting,
        
        onStartAddingPoint, isAddingPoint,
        onStartSettingPosB, isSettingPosB,
        
        // 用于 SMART_TOGGLE 特定索引的点重设
        onStartSettingSmartToggleIndex, isSettingSmartToggleIndex,
        
        getRemoteApiImg, openFileInput,
        
        visualSwitches, setVisualSwitches,
        setTempImg,

        currentActiveKey, 
        addingMacroPoint, setAddingMacroPoint,
        
        onMouseBtnClick
    } = props;

    const [continuousShot, setContinuousShot] = useState(false);
    const [intervalMs, setIntervalMs] = useState(500);
    const timerRef = useRef(null);
    const [infoBadgeVisible, setInfoBadgeVisible] = useState(true);

    const toggleContinuousShot = () => {
        if (continuousShot) {
            setContinuousShot(false);
            if (timerRef.current) {
                clearTimeout(timerRef.current);
                timerRef.current = null;
            }
            const url = `/screen.png?t=${Date.now()}`;
            getRemoteApiImg(url); 
        } else {
            setContinuousShot(true);
            const fetchLoop = () => {
                const url = `/screen.png?t=${Date.now()}`;
                const img = new Image();
                img.onload = () => {
                    if (timerRef.current) {
                        setTempImg(url);
                        timerRef.current = setTimeout(fetchLoop, intervalMs);
                    }
                };
                img.onerror = () => {
                    console.error("截图加载失败");
                    if (timerRef.current) {
                        timerRef.current = setTimeout(fetchLoop, intervalMs);
                    }
                };
                img.src = url;
            };
            fetchLoop();
            timerRef.current = 1; 
        }
    };

    useEffect(() => {
        return () => {
            if (timerRef.current) clearTimeout(timerRef.current);
        };
    }, []);

    const [activeTab, setActiveTab] = useState("GENERAL");

    const tabButtonStyle = (tabName) => ({
        justifyContent: "flex-start",
        paddingLeft: "10px",
        textTransform: "none",
        color: activeTab === tabName ? "#fff" : "#555",
        backgroundColor: activeTab === tabName ? "#009688" : "transparent",
        '&:hover': { backgroundColor: activeTab === tabName ? "#00796B" : "#eee" },
        borderRadius: "0px",
        width: "100%",
        minHeight: "40px",
        minWidth: "40px"
    });

    return (
        <Paper elevation={3} sx={{ height: '450px', width: '330px', display: 'flex', overflow: 'hidden', borderRadius: 2 }}>
            <Box sx={{ width: '50px', bgcolor: '#e0e0e0', borderRight: '1px solid #ccc', display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                <Button sx={tabButtonStyle("GENERAL")} onClick={() => setActiveTab("GENERAL")}><SettingsIcon /></Button>
                <Button sx={tabButtonStyle("KEYS")} onClick={() => setActiveTab("KEYS")}><KeyboardIcon /></Button>
                <Button sx={tabButtonStyle("COMBO")} onClick={() => setActiveTab("COMBO")}><LinkIcon /></Button>
                <Button sx={tabButtonStyle("MACRO")} onClick={() => setActiveTab("MACRO")}><MovieFilterIcon /></Button>
                <Button sx={tabButtonStyle("VIEW")} onClick={() => setActiveTab("VIEW")}><ViewInArIcon /></Button>
                <Button sx={tabButtonStyle("VMOUSE")} onClick={() => setActiveTab("VMOUSE")}><MouseIcon /></Button>
                <Button sx={tabButtonStyle("MAP")} onClick={() => setActiveTab("MAP")}><MapIcon /></Button>
                <Button sx={tabButtonStyle("WHEEL")} onClick={() => setActiveTab("WHEEL")}><RadioButtonCheckedIcon /></Button>
                <Button sx={tabButtonStyle("SCROLL")} onClick={() => setActiveTab("SCROLL")}><HeightIcon /></Button>
                <Button sx={tabButtonStyle("RECOIL")} onClick={() => setActiveTab("RECOIL")}><GpsFixedIcon /></Button>
                {Object.keys(pluginConfig).length > 0 && <Button sx={tabButtonStyle("PLUGIN")} onClick={() => setActiveTab("PLUGIN")}><ExtensionIcon /></Button>}
            </Box>
            <Box sx={{ flex: 1, overflowY: 'auto', bgcolor: '#F5F5F5', position: 'relative' }}>
                <Box sx={{ p: 1, borderBottom: '1px solid #ddd', bgcolor: '#fff', display: 'flex', justifyContent: 'space-between', alignItems: 'center', position: 'sticky', top: 0, zIndex: 10 }}>
                    <Typography variant="caption" sx={{ fontWeight: 'bold', fontSize: '0.7rem', color: '#1976D2' }}>
                        {currentActiveKey ? `当前按下: ${currentActiveKey}` : "按住键盘/手柄按键..."}
                    </Typography>
                    <Button size="small" variant="contained" color={isExporting ? "secondary" : "primary"} onClick={exportJSON} sx={{ fontSize: '0.65rem', padding: "2px 8px" }}>{exportButtonText}</Button>
                </Box>
                
                {activeTab === "GENERAL" && <GeneralSettings 
                    config={config} setConfig={setConfig} 
                    getRemoteApiImg={getRemoteApiImg} openFileInput={openFileInput}
                    continuousShot={continuousShot} toggleContinuousShot={toggleContinuousShot}
                    intervalMs={intervalMs} setIntervalMs={setIntervalMs}
                    infoBadgeVisible={infoBadgeVisible} setInfoBadgeVisible={setInfoBadgeVisible}
                    currentActiveKey={currentActiveKey} 
                />}
                
                {activeTab === "KEYS" && <KeysSettings 
                    config={config} setConfig={setConfig} setSelectKEY={setSelectKEY} 
                    visualSwitches={visualSwitches} setVisualSwitches={setVisualSwitches}
                    onStartAddingPoint={onStartAddingPoint} isAddingPoint={isAddingPoint}
                    onStartSettingPosB={onStartSettingPosB} isSettingPosB={isSettingPosB}
                    onStartSettingSmartToggleIndex={onStartSettingSmartToggleIndex} isSettingSmartToggleIndex={isSettingSmartToggleIndex}
                    onMouseBtnClick={onMouseBtnClick}
                />}
                
                {activeTab === "COMBO" && <ComboSettings 
                    config={config} setConfig={setConfig} 
                    visualSwitches={visualSwitches} setVisualSwitches={setVisualSwitches}
                    onStartAddingPoint={onStartAddingPoint} isAddingPoint={isAddingPoint}
                    onStartSettingPosB={onStartSettingPosB} isSettingPosB={isSettingPosB}
                    onStartSettingSmartToggleIndex={onStartSettingSmartToggleIndex} 
                    isSettingSmartToggleIndex={isSettingSmartToggleIndex}
                />}

                {activeTab === "MACRO" && <MacroEditor 
                    config={config} setConfig={setConfig}
                    currentActiveKey={currentActiveKey}
                    visualSwitches={visualSwitches} setVisualSwitches={setVisualSwitches}
                    addingMacroPoint={addingMacroPoint} setAddingMacroPoint={setAddingMacroPoint}
                />}

                {activeTab === "VIEW" && <ViewSettings 
                    config={config} setConfig={setConfig} 
                    setViewCenterSetting={setViewCenterSetting} isViewCenterSetting={isViewCenterSetting} 
                    visualSwitches={visualSwitches} setVisualSwitches={setVisualSwitches}
                />}
                
                {activeTab === "VMOUSE" && <VMouseSettings 
                    config={config} setConfig={setConfig} 
                    setVMouseResetPosSetting={setVMouseResetPosSetting} isVMouseResetPosSetting={isVMouseResetPosSetting}
                    visualSwitches={visualSwitches} setVisualSwitches={setVisualSwitches}
                    currentActiveKey={currentActiveKey}
                />}
                
                {activeTab === "MAP" && <MapSettings 
                    config={config} setConfig={setConfig} 
                    currentActiveKey={currentActiveKey}
                />}
                
                {activeTab === "WHEEL" && <WheelSettings 
                    config={config} setConfig={setConfig} 
                    setWheelPosSelecting={setWheelPosSelecting} isWheelPosSelecting={isWheelPosSelecting}
                    visualSwitches={visualSwitches} setVisualSwitches={setVisualSwitches}
                />}
                
                {activeTab === "SCROLL" && <ScrollSettings 
                    config={config} setConfig={setConfig} 
                    setScrollSliderPosSelecting={setScrollSliderPosSelecting} isScrollSliderPosSelecting={isScrollSliderPosSelecting}
                    visualSwitches={visualSwitches} setVisualSwitches={setVisualSwitches}
                />}
                
                {activeTab === "RECOIL" && <RecoilSettings 
                    config={config} setConfig={setConfig} 
                    currentActiveKey={currentActiveKey}
                />}
                
                {activeTab === "PLUGIN" && <PluginSettings 
                    pluginConfig={pluginConfig} pluginValue={pluginValue} setPluginValue={setPluginValue} 
                />}
            </Box>
        </Paper>
    );
}