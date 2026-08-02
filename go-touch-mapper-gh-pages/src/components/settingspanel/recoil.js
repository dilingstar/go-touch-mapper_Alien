import React from 'react';
import { Grid, Typography, Switch, Button, Box } from "@mui/material";
import HighlightOffIcon from '@mui/icons-material/HighlightOff';
import { produce } from "immer";
import { CostumedInput } from "../UIcomponents";

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

// [V3.4.2] 恢复鼠标键按钮，同时支持按键绑定
export default function RecoilSettings({ config, setConfig, currentActiveKey }) {
    
    const KeyBinder = ({ label, targetKeyArr }) => {
        const mouseBtns = ["BTN_LEFT", "BTN_RIGHT", "BTN_MIDDLE", "BTN_SIDE", "BTN_EXTRA"];
        const currentKeys = config.GLOBAL_RECOIL[targetKeyArr] || [];
        
        const removeKey = (idx) => setConfig(produce(d => {
            d.GLOBAL_RECOIL[targetKeyArr].splice(idx, 1);
        }));

        const toggleKey = (key) => setConfig(produce(d => {
            const arr = d.GLOBAL_RECOIL[targetKeyArr];
            const idx = arr.indexOf(key);
            if(idx !== -1) arr.splice(idx, 1); else arr.push(key);
        }));

        const addCurrentKey = () => {
            if (currentActiveKey) {
                setConfig(produce(d => {
                    const arr = d.GLOBAL_RECOIL[targetKeyArr];
                    if (arr.indexOf(currentActiveKey) === -1) {
                        arr.push(currentActiveKey);
                    }
                }));
            }
        };

        return (
            <Box sx={{ mt: 1, p: 1, bgcolor: '#f0f0f0', borderRadius: 1 }}>
                <Typography variant="caption">{label}</Typography>
                
                {/* 鼠标快捷键 */}
                <Grid container spacing={0.5} sx={{ mt: 0.5 }}>
                    {mouseBtns.map(btn => (
                        <Grid item key={btn}>
                            <Button variant={currentKeys.includes(btn) ? "contained" : "outlined"} size="small" sx={{ minWidth: "30px", px: 1, fontSize: '0.65rem' }} onClick={() => toggleKey(btn)}>{btn.replace("BTN_", "")}</Button>
                        </Grid>
                    ))}
                </Grid>

                {/* 绑定按钮 */}
                <Grid container spacing={0.5} sx={{ mt: 0.5 }}>
                    <Grid item>
                        <Button 
                            variant="outlined" size="small" 
                            sx={{ minWidth: "30px", px: 1, fontSize: '0.65rem', borderColor: currentActiveKey ? '#ff9800' : '#ccc', color: currentActiveKey ? '#e65100' : '#999' }} 
                            onClick={addCurrentKey}
                            disabled={!currentActiveKey}
                        >
                            {currentActiveKey ? `绑定: ${currentActiveKey}` : "按住按键绑定"}
                        </Button>
                    </Grid>
                </Grid>

                <Box sx={{ display: 'flex', flexWrap: 'wrap', mt: 0.5 }}>
                    {currentKeys.filter(k => !mouseBtns.includes(k)).map((k, idx) => (
                        <Typography key={idx} variant="caption" sx={{ border: '1px solid #999', borderRadius: 1, px: 0.5, mr: 0.5, mb: 0.5, display: 'flex', alignItems: 'center', bgcolor: '#fff' }}>
                            {k.replace("BTN_", "").replace("KEY_", "")} 
                            <HighlightOffIcon sx={{ fontSize: '0.8rem', cursor: 'pointer', ml: 0.5, color: '#f44336' }} onClick={() => toggleKey(k)} />
                        </Typography>
                    ))}
                </Box>
            </Box>
        );
    };

    return (
        <div style={{ padding: '8px' }}>
            <Typography variant="subtitle2" sx={{color: "#F44336", mb: 1, fontWeight: 'bold', fontSize: '0.8rem'}}>全局压枪</Typography>
            <Grid container justifyContent="space-between" alignItems="center">
                <Typography variant="body2" sx={{fontSize: '0.75rem'}}>启用压枪</Typography>
                <Switch size="small" checked={config.GLOBAL_RECOIL.ENABLE} onChange={() => setConfig(produce(d => { d.GLOBAL_RECOIL.ENABLE = !d.GLOBAL_RECOIL.ENABLE }))} />
            </Grid>
            <KeyBinder label="触发键 (Fire):" targetKeyArr="TRIGGER_KEYS" />
            <Grid container justifyContent="space-between" alignItems="center" sx={{ mt: 1 }}>
                <Typography variant="body2" sx={{fontSize: '0.75rem'}}>开镜模式</Typography>
                <Switch size="small" checked={config.GLOBAL_RECOIL.SCOPE_MODE} onChange={() => setConfig(produce(d => { d.GLOBAL_RECOIL.SCOPE_MODE = !d.GLOBAL_RECOIL.SCOPE_MODE }))} />
            </Grid>
            <KeyBinder label="开镜键 (Scope):" targetKeyArr="SCOPE_KEYS" />
            <Typography variant="body2" sx={{ mt: 1, fontSize: '0.75rem' }}>基础下压速度 (0-100):</Typography>
            <SliderWithInput label="" value={config.GLOBAL_RECOIL.BASE_SPEED} min={0} max={100} step={1} width="50px" onChange={v => setConfig(produce(d => { d.GLOBAL_RECOIL.BASE_SPEED = Number(v) }))} />
            <KeyBinder label="重置速度键:" targetKeyArr="RESET_SPEED_KEYS" />
        </div>
    );
}