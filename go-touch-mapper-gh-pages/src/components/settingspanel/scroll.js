import React from 'react';
import { Grid, Typography, Switch, Button } from "@mui/material";
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

export default function ScrollSettings({ config, setConfig, setScrollSliderPosSelecting, isScrollSliderPosSelecting, visualSwitches, setVisualSwitches }) {
    return (
        <div style={{ padding: '8px' }}>
            <Typography variant="subtitle2" sx={{color: "#FF9800", mb: 1, fontWeight: 'bold', fontSize: '0.8rem'}}>滚轮滑块</Typography>
            
            {/* 可视化开关 */}
            <Grid container alignItems="center" sx={{ mb: 1, bgcolor: '#fff3e0', p: 0.5, borderRadius: 1 }}>
                <Typography variant="caption" sx={{ mr: 1 }}>滑块可视化:</Typography>
                <Switch size="small" checked={visualSwitches.scroll} onChange={() => setVisualSwitches(produce(d => { d.scroll = !d.scroll }))} />
            </Grid>

            <Grid container justifyContent="space-between" alignItems="center">
                <Typography variant="body2" sx={{fontSize: '0.75rem'}}>启用滑块</Typography>
                <Switch size="small" checked={config.SCROLL_SLIDER.ENABLE} onChange={() => setConfig(produce(d => { d.SCROLL_SLIDER.ENABLE = !d.SCROLL_SLIDER.ENABLE }))} />
            </Grid>
            <Grid container alignItems="center" sx={{ mb: 1 }}>
                <Typography variant="body2" sx={{fontSize: '0.75rem'}}>位置:</Typography>
                <Button size="small" variant="outlined" sx={{ ml: 1, fontSize: '0.65rem' }} onClick={setScrollSliderPosSelecting} disabled={isScrollSliderPosSelecting && isScrollSliderPosSelecting()}>重设位置</Button>
            </Grid>
            <SliderWithInput label="上滑%" value={config.SCROLL_SLIDER.LENGTH_UP * 100} min={1} max={50} step={0.5} onChange={v => setConfig(produce(d => { d.SCROLL_SLIDER.LENGTH_UP = Number(v) / 100 }))} />
            <SliderWithInput label="下滑%" value={config.SCROLL_SLIDER.LENGTH_DOWN * 100} min={1} max={50} step={0.5} onChange={v => setConfig(produce(d => { d.SCROLL_SLIDER.LENGTH_DOWN = Number(v) / 100 }))} />
            
            <Grid container alignItems="center" spacing={1} sx={{ mt: 0.5 }}>
                <Grid item xs={6}><Typography variant="caption">延迟释放ms:</Typography></Grid>
                <Grid item xs={6}><CostumedInput defaultValue={config.SCROLL_SLIDER.RELEASE_DELAY_MS} onCommit={v => setConfig(produce(d => { d.SCROLL_SLIDER.RELEASE_DELAY_MS = Number(v) }))} width="40px" /></Grid>
                <Grid item xs={6}><Typography variant="caption">顶底延迟ms:</Typography></Grid>
                <Grid item xs={6}><CostumedInput defaultValue={config.SCROLL_SLIDER.DELAY_RESET_MS || 20} onCommit={v => setConfig(produce(d => { d.SCROLL_SLIDER.DELAY_RESET_MS = Number(v) }))} width="40px" /></Grid>
                <Grid item xs={6}><Typography variant="caption">非重置s:</Typography></Grid>
                <Grid item xs={6}><CostumedInput defaultValue={config.SCROLL_SLIDER.TIMEOUT_S || 3.0} onCommit={v => setConfig(produce(d => { d.SCROLL_SLIDER.TIMEOUT_S = Number(v) }))} width="40px" /></Grid>
            </Grid>
            
            <SliderWithInput label="速度" value={config.SCROLL_SLIDER.SPEED} min={0.1} max={10} step={0.1} onChange={v => setConfig(produce(d => { d.SCROLL_SLIDER.SPEED = Number(v) }))} />
            
            <Typography variant="body2" sx={{ mt: 1, fontSize: '0.75rem' }}>算法设置:</Typography>
            <Grid container alignItems="center">
                <Typography variant="caption">随机范围</Typography>
                <Switch size="small" checked={config.SCROLL_SLIDER.RANDOM_START_ENABLE} onChange={() => setConfig(produce(d => { d.SCROLL_SLIDER.RANDOM_START_ENABLE = !d.SCROLL_SLIDER.RANDOM_START_ENABLE }))} />
            </Grid>
            <SliderWithInput label="半径%" disabled={!config.SCROLL_SLIDER.RANDOM_START_ENABLE} value={config.SCROLL_SLIDER.RANDOM_START_RADIUS * 100} min={0} max={5} step={0.1} onChange={v => setConfig(produce(d => { d.SCROLL_SLIDER.RANDOM_START_RADIUS = Number(v) / 100 }))} />
            <Grid container alignItems="center">
                <Typography variant="caption">曲线路径</Typography>
                <Switch size="small" checked={config.SCROLL_SLIDER.CURVE_ENABLE} onChange={() => setConfig(produce(d => { d.SCROLL_SLIDER.CURVE_ENABLE = !d.SCROLL_SLIDER.CURVE_ENABLE }))} />
            </Grid>
            <SliderWithInput label="幅度%" disabled={!config.SCROLL_SLIDER.CURVE_ENABLE} value={config.SCROLL_SLIDER.CURVE_AMOUNT * 100} min={0} max={2} step={0.1} onChange={v => setConfig(produce(d => { d.SCROLL_SLIDER.CURVE_AMOUNT = Number(v) / 100 }))} />
            <SliderWithInput label="频率" disabled={!config.SCROLL_SLIDER.CURVE_ENABLE} value={config.SCROLL_SLIDER.CURVE_FREQUENCY || 1.0} min={0.1} max={10} step={0.1} onChange={v => setConfig(produce(d => { d.SCROLL_SLIDER.CURVE_FREQUENCY = Number(v) }))} />
            
            <Grid container alignItems="center">
                <Typography variant="caption">动态噪点</Typography>
                <Switch size="small" checked={config.SCROLL_SLIDER.NOISE_ENABLE} onChange={() => setConfig(produce(d => { d.SCROLL_SLIDER.NOISE_ENABLE = !d.SCROLL_SLIDER.NOISE_ENABLE }))} />
            </Grid>
            <SliderWithInput label="强度" disabled={!config.SCROLL_SLIDER.NOISE_ENABLE} value={config.SCROLL_SLIDER.NOISE_INTENSITY * 1000} min={0} max={10} step={0.1} onChange={v => setConfig(produce(d => { d.SCROLL_SLIDER.NOISE_INTENSITY = Number(v) / 1000 }))} />
        </div>
    );
}