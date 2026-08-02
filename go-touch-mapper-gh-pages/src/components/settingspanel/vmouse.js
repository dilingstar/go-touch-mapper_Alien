import React from 'react';
import { Grid, Typography, Switch, Button, Divider, Box } from "@mui/material";
import HighlightOffIcon from '@mui/icons-material/HighlightOff';
import { produce } from "immer";
import { CostumedInput } from "../UIcomponents";

// Version: V3.4.5

const SliderWithInput = ({ label, value, onChange, min, max, step, disabled = false, width = "35px" }) => {
    const { Slider } = require("@mui/material");
    
    const handleSliderChange = (event, newValue) => { onChange(newValue); };
    const handleInputChange = (newValue) => {
        let numValue = parseFloat(newValue);
        if (isNaN(numValue)) return;
        if (numValue < min) numValue = min;
        if (numValue > max) numValue = max;
        onChange(numValue);
    };
    const safeValue = (typeof value === 'number' && !isNaN(value)) ? value : min;
    
    return (
        <Grid container direction="row" justifyContent="space-between" alignItems="center" spacing={0.5} sx={{ mt: 0, mb: 0, height: "28px" }}>
            <Grid item xs={4} sx={{ pr: 0 }}>
                <Typography gutterBottom sx={{ fontSize: "0.7rem", whiteSpace: "nowrap", color: disabled ? 'text.disabled' : 'text.primary' }}>
                    {label}
                </Typography>
            </Grid>
            <Grid item xs={5} sx={{ pl: 0, pr: 0 }}>
                <Slider disabled={disabled} min={min} max={max} step={step} value={safeValue} onChange={handleSliderChange} size="small" />
            </Grid>
            <Grid item xs={3} sx={{ pl: 0.5 }}>
                <CostumedInput key={safeValue + (disabled ? '-disabled' : '')} defaultValue={safeValue} onCommit={handleInputChange} width={width} disabled={disabled} type="number" inputProps={{ step: step }} />
            </Grid>
        </Grid>
    );
};

export default function VMouseSettings({ config, setConfig, setVMouseResetPosSetting, isVMouseResetPosSetting, visualSwitches, setVisualSwitches, currentActiveKey }) {
    
    // [V3.4.5] 等价键绑定组件
    const KeyBinder = ({ label, targetKeyArr }) => {
        const currentKeys = config.V_MOUSE_SETTINGS[targetKeyArr] || [];
        
        const removeKey = (idx) => setConfig(produce(d => {
            d.V_MOUSE_SETTINGS[targetKeyArr].splice(idx, 1);
        }));

        const addCurrentKey = () => {
            if (currentActiveKey) {
                setConfig(produce(d => {
                    if (!d.V_MOUSE_SETTINGS[targetKeyArr]) d.V_MOUSE_SETTINGS[targetKeyArr] = [];
                    const arr = d.V_MOUSE_SETTINGS[targetKeyArr];
                    if (arr.indexOf(currentActiveKey) === -1) {
                        arr.push(currentActiveKey);
                    }
                }));
            }
        };

        return (
            <Grid container alignItems="center" spacing={1} sx={{ mt: 0.5 }}>
                <Grid item xs={4}><Typography variant="caption">{label}</Typography></Grid>
                <Grid item xs={8}>
                    <Button 
                        variant="outlined" size="small" 
                        sx={{ minWidth: "30px", px: 1, fontSize: '0.65rem', borderColor: currentActiveKey ? '#ff9800' : '#ccc', color: currentActiveKey ? '#e65100' : '#999', width: '100%' }} 
                        onClick={addCurrentKey}
                        disabled={!currentActiveKey}
                    >
                        {currentActiveKey ? `绑定: ${currentActiveKey}` : "按住按键绑定"}
                    </Button>
                </Grid>
                {currentKeys.length > 0 && (
                    <Grid item xs={12}>
                        <Box sx={{ display: 'flex', flexWrap: 'wrap' }}>
                            {currentKeys.map((k, idx) => (
                                <Typography key={idx} variant="caption" sx={{ border: '1px solid #999', borderRadius: 1, px: 0.5, mr: 0.5, mb: 0.5, display: 'flex', alignItems: 'center', bgcolor: '#fff' }}>
                                    {k.replace("BTN_", "").replace("KEY_", "")} 
                                    <HighlightOffIcon sx={{ fontSize: '0.8rem', cursor: 'pointer', ml: 0.5, color: '#f44336' }} onClick={() => removeKey(idx)} />
                                </Typography>
                            ))}
                        </Box>
                    </Grid>
                )}
            </Grid>
        );
    };

    return (
        <div style={{ padding: '8px' }}>
            <Typography variant="subtitle2" sx={{color: "#00E676", mb: 1, fontWeight: 'bold', fontSize: '0.8rem'}}>-v 悬浮鼠标</Typography>
            
            {/* 可视化开关 */}
            <Grid container alignItems="center" sx={{ mb: 1, bgcolor: '#e8f5e9', p: 0.5, borderRadius: 1 }}>
                <Typography variant="caption" sx={{ mr: 1 }}>切出位置可视化:</Typography>
                <Switch size="small" checked={visualSwitches.vMouse} onChange={() => setVisualSwitches(produce(d => { d.vMouse = !d.vMouse }))} />
            </Grid>

            <Grid container alignItems="center" sx={{ mb: 1 }}>
                <Typography variant="body2" sx={{fontSize: '0.75rem'}}>切出位置:</Typography>
                <Button size="small" variant="outlined" sx={{ ml: 1, fontSize: '0.65rem', borderColor: "#00E676", color: "#00E676" }} onClick={setVMouseResetPosSetting} disabled={isVMouseResetPosSetting && isVMouseResetPosSetting()}>重设位置</Button>
            </Grid>

            <SliderWithInput label="X速度" value={config.V_MOUSE_SETTINGS.MOUSE_SPEED[0]} min={0.1} max={5} step={0.1} onChange={v => setConfig(produce(d => { d.V_MOUSE_SETTINGS.MOUSE_SPEED[0] = Number(v) }))} />
            <SliderWithInput label="Y速度" value={config.V_MOUSE_SETTINGS.MOUSE_SPEED[1]} min={0.1} max={5} step={0.1} onChange={v => setConfig(produce(d => { d.V_MOUSE_SETTINGS.MOUSE_SPEED[1] = Number(v) }))} />
            
            <Divider sx={{ my: 1 }} />
            
            <Typography variant="body2" sx={{fontSize: '0.75rem'}}>等价键配置 (解决手柄点击):</Typography>
            <Box sx={{ p: 1, bgcolor: '#f0f0f0', borderRadius: 1, mb: 1, mt: 0.5 }}>
                <KeyBinder label="左键等价:" targetKeyArr="LEFT_CLICK_KEYS" />
                <KeyBinder label="右键等价:" targetKeyArr="RIGHT_CLICK_KEYS" />
            </Box>

            <Divider sx={{ my: 1 }} />
            
            <Grid container justifyContent="space-between" alignItems="center">
                <Typography variant="body2" sx={{fontSize: '0.75rem'}}>滚轮配置:</Typography>
                <Grid item sx={{ display: 'flex', alignItems: 'center' }}>
                    <Typography variant="caption" sx={{ mr: 0.5 }}>反转滚轮</Typography>
                    <Switch size="small" checked={config.V_MOUSE_SETTINGS.ENABLE_INVERT_SCROLL} onChange={() => setConfig(produce(d => { d.V_MOUSE_SETTINGS.ENABLE_INVERT_SCROLL = !d.V_MOUSE_SETTINGS.ENABLE_INVERT_SCROLL }))} />
                </Grid>
            </Grid>
            <SliderWithInput label="速度" value={config.V_MOUSE_SETTINGS.SCROLL_CONFIG.SPEED} min={0.1} max={5} step={0.1} onChange={v => setConfig(produce(d => { d.V_MOUSE_SETTINGS.SCROLL_CONFIG.SPEED = Number(v) }))} />
            <Grid container alignItems="center" spacing={1} sx={{ mt: 0 }}>
                <Grid item xs={6}><Typography variant="caption">释放延迟:</Typography></Grid>
                <Grid item xs={6}><CostumedInput defaultValue={config.V_MOUSE_SETTINGS.SCROLL_CONFIG.RELEASE_DELAY_MS || 50} onCommit={v => setConfig(produce(d => { d.V_MOUSE_SETTINGS.SCROLL_CONFIG.RELEASE_DELAY_MS = Number(v) }))} width="40px" /></Grid>
                <Grid item xs={6}><Typography variant="caption">边缘重置:</Typography></Grid>
                <Grid item xs={6}><CostumedInput defaultValue={config.V_MOUSE_SETTINGS.SCROLL_CONFIG.RESET_DELAY_MS || 50} onCommit={v => setConfig(produce(d => { d.V_MOUSE_SETTINGS.SCROLL_CONFIG.RESET_DELAY_MS = Number(v) }))} width="40px" /></Grid>
                <Grid item xs={6}><Typography variant="caption">非重置:</Typography></Grid>
                <Grid item xs={6}><CostumedInput defaultValue={config.V_MOUSE_SETTINGS.SCROLL_CONFIG.NON_RESET_MS || 300} onCommit={v => setConfig(produce(d => { d.V_MOUSE_SETTINGS.SCROLL_CONFIG.NON_RESET_MS = Number(v) }))} width="40px" /></Grid>
            </Grid>
            <Grid container alignItems="center" sx={{ mt: 0.5 }}>
                <Typography variant="body2" sx={{ mr: 1, fontSize: '0.75rem' }}>路径曲线:</Typography>
                <Switch size="small" checked={config.V_MOUSE_SETTINGS.SCROLL_CONFIG.CURVE_ENABLE} onChange={() => setConfig(produce(d => { d.V_MOUSE_SETTINGS.SCROLL_CONFIG.CURVE_ENABLE = !d.V_MOUSE_SETTINGS.SCROLL_CONFIG.CURVE_ENABLE }))} />
            </Grid>
            <SliderWithInput label="幅度" disabled={!config.V_MOUSE_SETTINGS.SCROLL_CONFIG.CURVE_ENABLE} value={config.V_MOUSE_SETTINGS.SCROLL_CONFIG.CURVE_AMOUNT * 100} min={0} max={5} step={0.1} onChange={v => setConfig(produce(d => { d.V_MOUSE_SETTINGS.SCROLL_CONFIG.CURVE_AMOUNT = Number(v) / 100 }))} />
            <SliderWithInput label="频率" disabled={!config.V_MOUSE_SETTINGS.SCROLL_CONFIG.CURVE_ENABLE} value={config.V_MOUSE_SETTINGS.SCROLL_CONFIG.CURVE_FREQ || 1.0} min={0.1} max={10} step={0.1} onChange={v => setConfig(produce(d => { d.V_MOUSE_SETTINGS.SCROLL_CONFIG.CURVE_FREQ = Number(v) }))} />
            <Grid container alignItems="center" sx={{ mt: 0.5 }}>
                <Typography variant="body2" sx={{ mr: 1, fontSize: '0.75rem' }}>动态噪点:</Typography>
                <Switch size="small" checked={config.V_MOUSE_SETTINGS.SCROLL_CONFIG.DYNAMIC_NOISE_ENABLE} onChange={() => setConfig(produce(d => { d.V_MOUSE_SETTINGS.SCROLL_CONFIG.DYNAMIC_NOISE_ENABLE = !d.V_MOUSE_SETTINGS.SCROLL_CONFIG.DYNAMIC_NOISE_ENABLE }))} />
            </Grid>
            <SliderWithInput label="强度" disabled={!config.V_MOUSE_SETTINGS.SCROLL_CONFIG.DYNAMIC_NOISE_ENABLE} value={config.V_MOUSE_SETTINGS.SCROLL_CONFIG.DYNAMIC_NOISE_AMOUNT * 100} min={0} max={2} step={0.05} onChange={v => setConfig(produce(d => { d.V_MOUSE_SETTINGS.SCROLL_CONFIG.DYNAMIC_NOISE_AMOUNT = Number(v) / 100 }))} />
        </div>
    );
}