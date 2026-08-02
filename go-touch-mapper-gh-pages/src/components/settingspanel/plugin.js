import React from 'react';
import { Typography, Switch, FormControl, Select, MenuItem, Box } from "@mui/material";
import { produce } from "immer";
import { CostumedInput, CostumedStringInput } from "../UIcomponents";

// 局部定义 SliderWithInput
const SliderWithInput = ({ label, value, onChange, min, max, step, disabled = false, width = "35px" }) => {
    const { Slider, Grid } = require("@mui/material");
    
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

export default function PluginSettings({ pluginConfig, pluginValue, setPluginValue }) {
    return (
        <div style={{ padding: '8px' }}>
            <Typography variant="subtitle2" sx={{color: "#9C27B0", mb: 1, fontWeight: 'bold', fontSize: '0.8rem'}}>插件配置</Typography>
            {Object.keys(pluginConfig).length === 0 && <Typography variant="caption">未检测到插件</Typography>}
            {Object.keys(pluginConfig).map((key) => {
                const template = pluginConfig[key];
                return (
                    <Box key={key} sx={{ mb: 2, borderBottom: '1px dashed #ccc', pb: 1 }}>
                        <Typography variant="body2" sx={{fontWeight: 'bold', fontSize: '0.75rem'}}>{key}</Typography>
                        <Typography variant="caption" sx={{color: '#666', display: 'block'}}>{template.description}</Typography>
                        {template.type === "int32" && (
                            <SliderWithInput label="Value" value={pluginValue[key]} min={template.min} max={template.max} step={1} onChange={v => setPluginValue(produce(d => { d[key] = v }))} />
                        )}
                        {template.type === "bool" && (
                            <Switch checked={pluginValue[key]} onChange={() => setPluginValue(produce(d => { d[key] = !d[key] }))} />
                        )}
                        {template.type === "string" && (
                            <CostumedStringInput defaultValue={pluginValue[key]} onCommit={v => setPluginValue(produce(d => { d[key] = v }))} width="100%" />
                        )}
                        {template.type === "select" && (
                            <FormControl fullWidth size="small" sx={{mt: 1}}>
                                <Select value={pluginValue[key]} onChange={e => setPluginValue(produce(d => { d[key] = e.target.value }))}>
                                    {template.values.map((v, idx) => <MenuItem key={idx} value={idx}>{v}</MenuItem>)}
                                </Select>
                            </FormControl>
                        )}
                    </Box>
                );
            })}
        </div>
    );
}