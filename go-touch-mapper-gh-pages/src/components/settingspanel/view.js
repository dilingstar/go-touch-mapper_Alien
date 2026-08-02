import React from 'react';
import { Grid, Typography, Switch, Button, Slider } from "@mui/material";
import { produce } from "immer";
import { CostumedInput } from "../UIcomponents";

const SliderWithInput = ({ label, value, onChange, min, max, step, disabled = false, width = "35px" }) => {
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

export default function ViewSettings({ config, setConfig, setViewCenterSetting, isViewCenterSetting, visualSwitches, setVisualSwitches }) {
    return (
        <div style={{ padding: '8px' }}>
            <Typography variant="subtitle2" sx={{color: "#009688", mb: 1, fontWeight: 'bold', fontSize: '0.8rem'}}>视角设置</Typography>
            
            {/* 可视化开关 */}
            <Grid container alignItems="center" sx={{ mb: 1, bgcolor: '#e0f2f1', p: 0.5, borderRadius: 1 }}>
                <Typography variant="caption" sx={{ mr: 1 }}>中心可视化:</Typography>
                <Switch size="small" checked={visualSwitches.viewCenter} onChange={() => setVisualSwitches(produce(d => { d.viewCenter = !d.viewCenter }))} />
                <Typography variant="caption" sx={{ mr: 1, ml: 1 }}>半径可视化:</Typography>
                <Switch size="small" checked={visualSwitches.viewRadius} onChange={() => setVisualSwitches(produce(d => { d.viewRadius = !d.viewRadius }))} />
            </Grid>

            <Grid container alignItems="center" spacing={1} sx={{ mb: 1 }}>
                <Grid item><Typography variant="body2" sx={{fontSize: '0.75rem'}}>灵敏度X:</Typography></Grid>
                <Grid item><CostumedInput defaultValue={config.MOUSE.SPEED[0]} onCommit={v => setConfig(produce(d => { d.MOUSE.SPEED[0] = Number(v) }))} /></Grid>
                <Grid item><Typography variant="body2" sx={{fontSize: '0.75rem'}}>Y:</Typography></Grid>
                <Grid item><CostumedInput defaultValue={config.MOUSE.SPEED[1]} onCommit={v => setConfig(produce(d => { d.MOUSE.SPEED[1] = Number(v) }))} /></Grid>
                <Grid item><Button size="small" variant="outlined" sx={{fontSize: '0.65rem'}} onClick={setViewCenterSetting} disabled={isViewCenterSetting && isViewCenterSetting()}>重设中心</Button></Grid>
            </Grid>

            <Grid container alignItems="center" sx={{ mb: 0.5 }}>
                <Typography variant="body2" sx={{ mr: 1, fontSize: '0.75rem' }}>自动释放:</Typography>
                <Switch size="small" checked={config.MOUSE.VIEW_AUTO_RELEASE_ENABLE} onChange={() => setConfig(produce(d => { d.MOUSE.VIEW_AUTO_RELEASE_ENABLE = !d.MOUSE.VIEW_AUTO_RELEASE_ENABLE }))} />
                <CostumedInput disabled={!config.MOUSE.VIEW_AUTO_RELEASE_ENABLE} defaultValue={config.MOUSE.VIEW_AUTO_RELEASE_MS} onCommit={v => setConfig(produce(d => { d.MOUSE.VIEW_AUTO_RELEASE_MS = Number(v) }))} width="35px" />
                <Typography variant="caption" sx={{ ml: 0.5 }}>ms</Typography>
            </Grid>

            <Grid container alignItems="center">
                <Typography variant="body2" sx={{ mr: 1, fontSize: '0.75rem' }}>重置半径（归中）:</Typography>
                <Switch size="small" checked={config.MOUSE.VIEW_RESET_RADIUS_ENABLE} onChange={() => setConfig(produce(d => { d.MOUSE.VIEW_RESET_RADIUS_ENABLE = !d.MOUSE.VIEW_RESET_RADIUS_ENABLE }))} />
            </Grid>
            <SliderWithInput label="半径%" disabled={!config.MOUSE.VIEW_RESET_RADIUS_ENABLE} value={config.MOUSE.VIEW_RESET_RADIUS * 100} min={1} max={50} step={0.5} onChange={v => setConfig(produce(d => { d.MOUSE.VIEW_RESET_RADIUS = Number(v) / 100 }))} />
            <SliderWithInput label="厚度%" disabled={!config.MOUSE.VIEW_RESET_RADIUS_ENABLE} value={config.MOUSE.VIEW_RESET_RADIUS_THICKNESS * 100} min={0.1} max={5} step={0.1} onChange={v => setConfig(produce(d => { d.MOUSE.VIEW_RESET_RADIUS_THICKNESS = Number(v) / 100 }))} />

            <Grid container alignItems="center" sx={{ mt: 0.5 }}>
                <Typography variant="body2" sx={{ mr: 1, fontSize: '0.75rem' }}>随机重置:</Typography>
                <Switch size="small" checked={config.MOUSE.VIEW_RANDOM_RESET_ENABLE} onChange={() => setConfig(produce(d => { d.MOUSE.VIEW_RANDOM_RESET_ENABLE = !d.MOUSE.VIEW_RANDOM_RESET_ENABLE }))} />
            </Grid>
            <SliderWithInput label="范围%" disabled={!config.MOUSE.VIEW_RANDOM_RESET_ENABLE} value={config.MOUSE.VIEW_RANDOM_RESET_RADIUS * 100} min={0.1} max={5} step={0.1} onChange={v => setConfig(produce(d => { d.MOUSE.VIEW_RANDOM_RESET_RADIUS = Number(v) / 100 }))} />

            <Grid container alignItems="center" sx={{ mt: 0.5 }}>
                <Typography variant="body2" sx={{ mr: 1, fontSize: '0.75rem' }}>延迟重置:</Typography>
                <Switch size="small" checked={config.MOUSE.VIEW_DELAY_RESET_ENABLE} onChange={() => setConfig(produce(d => { d.MOUSE.VIEW_DELAY_RESET_ENABLE = !d.MOUSE.VIEW_DELAY_RESET_ENABLE }))} />
                <CostumedInput disabled={!config.MOUSE.VIEW_DELAY_RESET_ENABLE} defaultValue={config.MOUSE.VIEW_DELAY_RESET_MS} onCommit={v => setConfig(produce(d => { d.MOUSE.VIEW_DELAY_RESET_MS = Number(v) }))} width="35px" />
                <Typography variant="caption" sx={{ ml: 0.5 }}>ms</Typography>
            </Grid>
            <Grid container alignItems="center" sx={{ mt: 0, ml: 2 }}>
                <Typography variant="caption" sx={{ mr: 1 }}>随机延迟:</Typography>
                <Switch size="small" checked={config.MOUSE.VIEW_DELAY_RESET_RANDOM_ENABLE} onChange={() => setConfig(produce(d => { d.MOUSE.VIEW_DELAY_RESET_RANDOM_ENABLE = !d.MOUSE.VIEW_DELAY_RESET_RANDOM_ENABLE }))} disabled={!config.MOUSE.VIEW_DELAY_RESET_ENABLE} />
                <CostumedInput disabled={!config.MOUSE.VIEW_DELAY_RESET_ENABLE || !config.MOUSE.VIEW_DELAY_RESET_RANDOM_ENABLE} defaultValue={config.MOUSE.VIEW_DELAY_RESET_MIN_MS} onCommit={v => setConfig(produce(d => { d.MOUSE.VIEW_DELAY_RESET_MIN_MS = Number(v) }))} width="35px" />
                <Typography variant="caption" sx={{ ml: 0.5 }}>min</Typography>
            </Grid>
        </div>
    );
}