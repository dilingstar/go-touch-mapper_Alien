import React from 'react';
import { Grid, Button, Typography, Divider, Box } from "@mui/material";
import HighlightOffIcon from '@mui/icons-material/HighlightOff';
import { produce } from "immer";

export default function MapSettings({ config, setConfig, currentActiveKey }) {
    
    // 添加当前按下的键到 Switch Keys
    const addSwitchKey = () => {
        if (currentActiveKey) {
            setConfig(produce(d => {
                if (d.MOUSE.SWITCH_KEYS.indexOf(currentActiveKey) === -1) {
                    d.MOUSE.SWITCH_KEYS.push(currentActiveKey);
                }
            }));
        }
    };

    // 添加当前按下的键到 Pointer Keys
    const addPointerKey = () => {
        if (currentActiveKey) {
            setConfig(produce(d => {
                if (!d.MOUSE.POINTER_SWITCH_KEYS) d.MOUSE.POINTER_SWITCH_KEYS = [];
                if (d.MOUSE.POINTER_SWITCH_KEYS.indexOf(currentActiveKey) === -1) {
                    d.MOUSE.POINTER_SWITCH_KEYS.push(currentActiveKey);
                }
            }));
        }
    };

    // [V3.4.7] 恢复 V3.3.0 的快捷中键切换逻辑
    const toggleSwitchKey = (btn) => setConfig(produce(d => {
        const arr = d.MOUSE.SWITCH_KEYS;
        const idx = arr.indexOf(btn);
        if(idx !== -1) arr.splice(idx, 1); else arr.push(btn);
    }));
    
    const togglePointerKey = (btn) => setConfig(produce(d => {
        if (!d.MOUSE.POINTER_SWITCH_KEYS) d.MOUSE.POINTER_SWITCH_KEYS = [];
        const arr = d.MOUSE.POINTER_SWITCH_KEYS;
        const idx = arr.indexOf(btn);
        if(idx !== -1) arr.splice(idx, 1); else arr.push(btn);
    }));

    return (
        <Box sx={{ p: 1 }}>
            <Typography variant="subtitle2" sx={{color: "#2196F3", mb: 1, fontWeight: 'bold', fontSize: '0.8rem'}}>映射切换</Typography>
            
            <Typography variant="body2" gutterBottom sx={{fontSize: '0.75rem'}}>切换映射 (点击):</Typography>
            <Grid container spacing={1} sx={{ mb: 1 }}>
                {config.MOUSE.SWITCH_KEYS.map((key, idx) => (
                    <Grid item key={idx}>
                        <Button variant="outlined" size="small" endIcon={<HighlightOffIcon />} sx={{fontSize: '0.65rem'}} onClick={() => setConfig(produce(d => { d.MOUSE.SWITCH_KEYS.splice(idx, 1) }))}>
                            {key.replace("KEY_", "").replace("BTN_", "")}
                        </Button>
                    </Grid>
                ))}
                {/* 绑定按钮 */}
                <Grid item>
                    <Button 
                        variant="outlined" size="small" sx={{fontSize: '0.65rem', borderColor: currentActiveKey ? '#4CAF50' : '#ccc', color: currentActiveKey ? '#4CAF50' : '#999'}} 
                        onClick={addSwitchKey}
                    >
                        {currentActiveKey ? "绑定当前" : "请先按键"}
                    </Button>
                </Grid>
                {/* [V3.4.7] 恢复的快捷中键按钮 */}
                <Grid item>
                    <Button 
                        variant={config.MOUSE.SWITCH_KEYS.includes("BTN_MIDDLE") ? "contained" : "outlined"} 
                        size="small" 
                        sx={{fontSize: '0.65rem'}} 
                        onClick={() => toggleSwitchKey("BTN_MIDDLE")}
                    >
                        中键
                    </Button>
                </Grid>
            </Grid>
            
            <Divider sx={{ my: 1 }} />
            
            <Typography variant="body2" gutterBottom sx={{fontSize: '0.75rem'}}>按住切出光标:</Typography>
            <Grid container spacing={1}>
                {(config.MOUSE.POINTER_SWITCH_KEYS || []).map((key, idx) => (
                    <Grid item key={idx}>
                        <Button variant="outlined" size="small" endIcon={<HighlightOffIcon />} sx={{fontSize: '0.65rem'}} onClick={() => setConfig(produce(d => { d.MOUSE.POINTER_SWITCH_KEYS.splice(idx, 1) }))}>
                            {key.replace("KEY_", "").replace("BTN_", "")}
                        </Button>
                    </Grid>
                ))}
                {/* 绑定按钮 */}
                <Grid item>
                    <Button 
                        variant="outlined" size="small" sx={{fontSize: '0.65rem', borderColor: currentActiveKey ? '#4CAF50' : '#ccc', color: currentActiveKey ? '#4CAF50' : '#999'}} 
                        onClick={addPointerKey}
                    >
                        {currentActiveKey ? "绑定当前" : "请先按键"}
                    </Button>
                </Grid>
                {/* [V3.4.7] 恢复的快捷中键按钮 */}
                <Grid item>
                    <Button 
                        variant={config.MOUSE.POINTER_SWITCH_KEYS && config.MOUSE.POINTER_SWITCH_KEYS.includes("BTN_MIDDLE") ? "contained" : "outlined"} 
                        size="small" 
                        sx={{fontSize: '0.65rem'}} 
                        onClick={() => togglePointerKey("BTN_MIDDLE")}
                    >
                        中键
                    </Button>
                </Grid>
            </Grid>
        </Box>
    );
}