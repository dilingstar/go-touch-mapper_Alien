import React from 'react';
import { Grid, Typography, Switch, Button, Box } from "@mui/material";
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

const FormControlLabelToggle = ({ label, checked, onChange }) => (
    <Grid container alignItems="center" sx={{ mr: 1 }}>
        <Switch size="small" checked={checked} onChange={onChange} disabled={!onChange} />
        <Typography variant="caption">{label}</Typography>
    </Grid>
);

export default function WheelSettings({ config, setConfig, setWheelPosSelecting, isWheelPosSelecting, visualSwitches, setVisualSwitches }) {
    return (
        <div style={{ padding: '8px' }}>
            <Typography variant="subtitle2" sx={{color: "#3F51B5", mb: 1, fontWeight: 'bold', fontSize: '0.8rem'}}>高级轮盘</Typography>
            
            {/* 可视化开关 */}
            <Grid container alignItems="center" sx={{ mb: 1, bgcolor: '#e8eaf6', p: 0.5, borderRadius: 1 }}>
                <Typography variant="caption" sx={{ mr: 1 }}>轮盘可视化:</Typography>
                <Switch size="small" checked={visualSwitches.wheel} onChange={() => setVisualSwitches(produce(d => { d.wheel = !d.wheel }))} />
            </Grid>

            <Grid container alignItems="center" sx={{ mb: 1 }}>
                <Typography variant="body2" sx={{fontSize: '0.75rem'}}>中心位置:</Typography>
                <Button size="small" variant="outlined" sx={{ ml: 1, fontSize: '0.65rem' }} onClick={setWheelPosSelecting} disabled={isWheelPosSelecting && isWheelPosSelecting()}>
                    {`(${parseInt(config.WHEEL.POS[0] * config.SCREEN.SIZE[0])},${parseInt(config.WHEEL.POS[1] * config.SCREEN.SIZE[1])})`}
                </Button>
            </Grid>
            <SliderWithInput label="半径%" value={config.WHEEL.RANGE * 100} min={1} max={50} step={0.5} onChange={v => setConfig(produce(d => { d.WHEEL.RANGE = Number(v) / 100 }))} />
            <Grid container alignItems="center" sx={{ mt: 0.5 }}>
                <Typography variant="body2" sx={{ mr: 1, fontSize: '0.75rem' }}>Shift加速:</Typography>
                <Switch size="small" checked={config.WHEEL.SHIFT_RANGE_ENABLE} onChange={() => setConfig(produce(d => { d.WHEEL.SHIFT_RANGE_ENABLE = !d.WHEEL.SHIFT_RANGE_ENABLE }))} />
            </Grid>
            <SliderWithInput label="Shift半径%" disabled={!config.WHEEL.SHIFT_RANGE_ENABLE} value={config.WHEEL.SHIFT_RANGE * 100} min={1} max={50} step={0.5} onChange={v => setConfig(produce(d => { d.WHEEL.SHIFT_RANGE = Number(v) / 100 }))} />
            
            <Grid container alignItems="center" sx={{ mt: 0.5, mb: 0.5 }}>
                <FormControlLabelToggle label="按下置换" checked={config.WHEEL.SHIFT_PRESS_TOGGLE} onChange={() => setConfig(produce(d => { d.WHEEL.SHIFT_PRESS_TOGGLE = !d.WHEEL.SHIFT_PRESS_TOGGLE }))} />
                <FormControlLabelToggle label="抬起置换" checked={config.WHEEL.SHIFT_RELEASE_TOGGLE} onChange={() => setConfig(produce(d => { d.WHEEL.SHIFT_RELEASE_TOGGLE = !d.WHEEL.SHIFT_RELEASE_TOGGLE }))} />
            </Grid>

            {/* [V3.2.5] 新增临时穿透开关 */}
            <Grid container alignItems="center" sx={{ mt: 0.5, mb: 0.5 }}>
                <Typography variant="body2" sx={{ mr: 1, fontSize: '0.75rem' }}>临时穿透:</Typography>
                <Switch size="small" checked={config.WHEEL.TEMP_PENETRATION_ENABLE || false} onChange={() => setConfig(produce(d => { d.WHEEL.TEMP_PENETRATION_ENABLE = !d.WHEEL.TEMP_PENETRATION_ENABLE }))} />
            </Grid>

            <Grid container alignItems="center" sx={{ mb: 0.5 }}>
                <Typography variant="body2" sx={{ mr: 1, fontSize: '0.75rem' }}>延迟释放:</Typography>
                <CostumedInput defaultValue={config.WHEEL.DELAY_RESET_MS} onCommit={v => setConfig(produce(d => { d.WHEEL.DELAY_RESET_MS = Number(v) }))} width="40px" />
                <Typography variant="caption" sx={{ ml: 0.5 }}>ms</Typography>
            </Grid>

            <Box sx={{ border: '1px solid #ddd', borderRadius: 1, p: 0.5, mt: 1, bgcolor: '#f9f9f9' }}>
                <Grid container alignItems="center" justifyContent="space-between">
                    <Typography variant="subtitle2" sx={{ color: "#D81B60", fontWeight: 'bold', fontSize: '0.75rem' }}>贝塞尔算法</Typography>
                    <Switch size="small" checked={config.WHEEL.BEZIER_ENABLE} onChange={() => setConfig(produce(d => { d.WHEEL.BEZIER_ENABLE = !d.WHEEL.BEZIER_ENABLE }))} />
                </Grid>
                <SliderWithInput label="移动速度" disabled={!config.WHEEL.BEZIER_ENABLE} value={config.WHEEL.BEZIER_SPEED} min={10} max={300} step={1} onChange={v => setConfig(produce(d => { d.WHEEL.BEZIER_SPEED = Number(v) }))} />
                {!config.WHEEL.BEZIER_ENABLE && 
                    <SliderWithInput label="线性平滑度" value={config.WHEEL.STEP_SPEED || 60} min={10} max={120} step={1} onChange={v => setConfig(produce(d => { d.WHEEL.STEP_SPEED = Number(v) }))} />
                }
                
                <SliderWithInput label="曲线幅度" disabled={!config.WHEEL.BEZIER_ENABLE} value={config.WHEEL.BEZIER_CURVE_AMOUNT} min={0} max={3} step={0.1} onChange={v => setConfig(produce(d => { d.WHEEL.BEZIER_CURVE_AMOUNT = Number(v) }))} />
                <SliderWithInput label="动态曲线" disabled={!config.WHEEL.BEZIER_ENABLE} value={config.WHEEL.BEZIER_DYNAMIC_CURVE} min={0} max={2} step={0.1} onChange={v => setConfig(produce(d => { d.WHEEL.BEZIER_DYNAMIC_CURVE = Number(v) }))} />

                <Grid container alignItems="center" sx={{ mt: 0.5 }}>
                    <Typography variant="body2" sx={{ mr: 1, fontSize: '0.75rem' }}>随机起终点:</Typography>
                    <Switch size="small" checked={config.WHEEL.RANDOM_POINT_ENABLE} onChange={() => setConfig(produce(d => { d.WHEEL.RANDOM_POINT_ENABLE = !d.WHEEL.RANDOM_POINT_ENABLE }))} disabled={!config.WHEEL.BEZIER_ENABLE} />
                </Grid>
                <SliderWithInput label="起点半径%" disabled={!config.WHEEL.BEZIER_ENABLE || !config.WHEEL.RANDOM_POINT_ENABLE} value={config.WHEEL.RANDOM_START_RADIUS * 100} min={0} max={10} step={0.1} onChange={v => setConfig(produce(d => { d.WHEEL.RANDOM_START_RADIUS = Number(v) / 100 }))} />
                <SliderWithInput label="终点半径%" disabled={!config.WHEEL.BEZIER_ENABLE || !config.WHEEL.RANDOM_POINT_ENABLE} value={config.WHEEL.RANDOM_END_RADIUS * 100} min={0} max={10} step={0.1} onChange={v => setConfig(produce(d => { d.WHEEL.RANDOM_END_RADIUS = Number(v) / 100 }))} />
                <SliderWithInput label="Shift终点%" disabled={!config.WHEEL.BEZIER_ENABLE || !config.WHEEL.RANDOM_POINT_ENABLE} value={(config.WHEEL.RANDOM_SHIFT_END_RADIUS || 0.015) * 100} min={0} max={10} step={0.1} onChange={v => setConfig(produce(d => { d.WHEEL.RANDOM_SHIFT_END_RADIUS = Number(v) / 100 }))} />

                <Grid container alignItems="center" sx={{ mt: 0.5 }}>
                    <Typography variant="body2" sx={{ mr: 1, fontSize: '0.75rem' }}>淡入淡出:</Typography>
                    <Switch size="small" checked={config.WHEEL.EASING_ENABLE} onChange={() => setConfig(produce(d => { d.WHEEL.EASING_ENABLE = !d.WHEEL.EASING_ENABLE }))} disabled={!config.WHEEL.BEZIER_ENABLE} />
                </Grid>
                <SliderWithInput label="淡入强度" disabled={!config.WHEEL.BEZIER_ENABLE || !config.WHEEL.EASING_ENABLE} value={config.WHEEL.EASING_IN} min={0} max={10} step={0.1} onChange={v => setConfig(produce(d => { d.WHEEL.EASING_IN = Number(v) }))} />
                <SliderWithInput label="淡出强度" disabled={!config.WHEEL.BEZIER_ENABLE || !config.WHEEL.EASING_ENABLE} value={config.WHEEL.EASING_OUT} min={0} max={10} step={0.1} onChange={v => setConfig(produce(d => { d.WHEEL.EASING_OUT = Number(v) }))} />

                <Grid container alignItems="center" sx={{ mt: 0.5 }}>
                    <Typography variant="body2" sx={{ mr: 1, fontSize: '0.75rem' }}>动态波噪点:</Typography>
                    <Switch size="small" checked={config.WHEEL.NOISE_ENABLE} onChange={() => setConfig(produce(d => { d.WHEEL.NOISE_ENABLE = !d.WHEEL.NOISE_ENABLE }))} disabled={!config.WHEEL.BEZIER_ENABLE} />
                </Grid>
                <SliderWithInput label="扰动强度" disabled={!config.WHEEL.BEZIER_ENABLE || !config.WHEEL.NOISE_ENABLE} value={config.WHEEL.NOISE_INTENSITY * 1000} min={0} max={20} step={0.5} onChange={v => setConfig(produce(d => { d.WHEEL.NOISE_INTENSITY = Number(v) / 1000 }))} />
            </Box>

            <Box sx={{ border: '1px solid #ddd', borderRadius: 1, p: 0.5, mt: 1 }}>
                <Grid container alignItems="center" justifyContent="space-between">
                    <Typography variant="subtitle2" sx={{ color: "#7B1FA2", fontSize: '0.75rem' }}>行星转圈</Typography>
                    <Switch size="small" checked={config.WHEEL.WHEEL_PLANET.ENABLE} onChange={() => setConfig(produce(d => { d.WHEEL.WHEEL_PLANET.ENABLE = !d.WHEEL.WHEEL_PLANET.ENABLE }))} />
                </Grid>
                <SliderWithInput label="半径" disabled={!config.WHEEL.WHEEL_PLANET.ENABLE} value={config.WHEEL.WHEEL_PLANET.RADIUS * 100} min={0} max={5} step={0.1} onChange={v => setConfig(produce(d => { d.WHEEL.WHEEL_PLANET.RADIUS = Number(v) / 100 }))} />
                <SliderWithInput label="速度" disabled={!config.WHEEL.WHEEL_PLANET.ENABLE} value={config.WHEEL.WHEEL_PLANET.SPEED} min={0.1} max={10} step={0.1} onChange={v => setConfig(produce(d => { d.WHEEL.WHEEL_PLANET.SPEED = Number(v) }))} />
                <SliderWithInput label="扰动强度" disabled={!config.WHEEL.WHEEL_PLANET.ENABLE} value={config.WHEEL.WHEEL_PLANET.PLANET_NOISE_INTENSITY * 1000 || 0} min={0} max={10} step={0.1} onChange={v => setConfig(produce(d => { d.WHEEL.WHEEL_PLANET.PLANET_NOISE_INTENSITY = Number(v) / 1000 }))} />
                
                <Grid container alignItems="center" sx={{ mt: 0.5 }}>
                    <Typography variant="caption" sx={{ mr: 1 }}>动态速度:</Typography>
                    <Switch size="small" checked={config.WHEEL.WHEEL_PLANET.PLANET_DYNAMIC_SPEED.ENABLE} onChange={() => setConfig(produce(d => { d.WHEEL.WHEEL_PLANET.PLANET_DYNAMIC_SPEED.ENABLE = !d.WHEEL.WHEEL_PLANET.PLANET_DYNAMIC_SPEED.ENABLE }))} disabled={!config.WHEEL.WHEEL_PLANET.ENABLE} />
                </Grid>
                <SliderWithInput label="最小速度" disabled={!config.WHEEL.WHEEL_PLANET.ENABLE || !config.WHEEL.WHEEL_PLANET.PLANET_DYNAMIC_SPEED.ENABLE} value={config.WHEEL.WHEEL_PLANET.PLANET_DYNAMIC_SPEED.MIN_SPEED} min={0} max={5} step={0.1} onChange={v => setConfig(produce(d => { d.WHEEL.WHEEL_PLANET.PLANET_DYNAMIC_SPEED.MIN_SPEED = Number(v) }))} />
                <SliderWithInput label="频率" disabled={!config.WHEEL.WHEEL_PLANET.ENABLE || !config.WHEEL.WHEEL_PLANET.PLANET_DYNAMIC_SPEED.ENABLE} value={config.WHEEL.WHEEL_PLANET.PLANET_DYNAMIC_SPEED.FREQUENCY} min={0.1} max={10} step={0.1} onChange={v => setConfig(produce(d => { d.WHEEL.WHEEL_PLANET.PLANET_DYNAMIC_SPEED.FREQUENCY = Number(v) }))} />
                
                <Grid container alignItems="center" sx={{ mt: 0.5 }}>
                    <Typography variant="caption" sx={{ mr: 1 }}>行星曲线:</Typography>
                    <Switch size="small" checked={config.WHEEL.PLANET_CURVE.ENABLE} onChange={() => setConfig(produce(d => { d.WHEEL.PLANET_CURVE.ENABLE = !d.WHEEL.PLANET_CURVE.ENABLE }))} disabled={!config.WHEEL.WHEEL_PLANET.ENABLE} />
                </Grid>
                <SliderWithInput label="幅度" disabled={!config.WHEEL.WHEEL_PLANET.ENABLE || !config.WHEEL.PLANET_CURVE.ENABLE} value={config.WHEEL.PLANET_CURVE.CURVE_AMOUNT * 100} min={0} max={5} step={0.1} onChange={v => setConfig(produce(d => { d.WHEEL.PLANET_CURVE.CURVE_AMOUNT = Number(v) / 100 }))} />
                <SliderWithInput label="频率" disabled={!config.WHEEL.WHEEL_PLANET.ENABLE || !config.WHEEL.PLANET_CURVE.ENABLE} value={config.WHEEL.PLANET_CURVE.CURVE_FREQUENCY || 1.0} min={0.1} max={10} step={0.1} onChange={v => setConfig(produce(d => { d.WHEEL.PLANET_CURVE.CURVE_FREQUENCY = Number(v) }))} />
            </Box>
        </div>
    );
}