import React from 'react';
import { Grid, Button, Typography, Switch, IconButton, FormControl, Select, MenuItem, Box, Slider } from "@mui/material";
import HighlightOffIcon from '@mui/icons-material/HighlightOff';
import AddLocationIcon from '@mui/icons-material/AddLocation';
import RestartAltIcon from '@mui/icons-material/RestartAlt';
import { produce } from "immer";
import { CostumedInput, CostumedStringInput } from "../UIcomponents";

// Version: V3.4.6

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

const KeyMapSettings = ({ data, config, setConfig, isCombo, onStartAddingPoint, isAddingPoint, onStartSettingPosB, isSettingPosB, onStartSettingSmartToggleIndex, isSettingSmartToggleIndex }) => {
    const handleChange = (e) => {
        const type = e.target.value;
        const currentPos = config.KEY_MAPS[data.KEY].POS || [0.5, 0.5];
        let newConfig = { TYPE: type, POS: currentPos };
        
        if (["AUTO_FIRE"].includes(type)) newConfig.INTERVAL = [18, 20, 10, 10];
        if (type === "CLICK") newConfig.INTERVAL = [18];
        if (["MULT_PRESS", "DRAG", "SEQUENTIAL_PRESS"].includes(type)) newConfig.POS_S = [];
        if (["BACKPACK_TOGGLE", "SYNC_BACKPACK", "PRESS_RELEASE_CLICK"].includes(type)) newConfig.POS_B = currentPos;
        if (type === "SYNC_BACKPACK" || type === "PRESS_RELEASE_CLICK" || type === "SEQUENTIAL_PRESS") newConfig.CLICK_DURATION = 50;
        if (type === "SEQUENTIAL_PRESS") newConfig.COOLDOWN = 0;
        
        // [V3.4.6 核心修复] 继承当前位置，避免飞到左上角
        if (type === "SMART_TOGGLE") {
            newConfig = {
                TYPE: "SMART_TOGGLE",
                POS_S: [currentPos, [currentPos[0] + 0.05, currentPos[1] + 0.05]],
                RELEASE_MOUSE: true,
                SEPARAT: false,
                TOUCH: true
            };
        }

        setConfig(produce(d => { d.KEY_MAPS[data.KEY] = newConfig }));
    };
    
    const isWheel = ["REL_WHEEL_UP", "REL_WHEEL_DOWN"].includes(data.KEY);

    const ensureSmartToggleInit = () => {
        if (data.TYPE === "SMART_TOGGLE") {
            if (!Array.isArray(data.POS_S) || data.POS_S.length < 2) {
                setConfig(produce(draft => {
                    const cur = draft.KEY_MAPS[data.KEY];
                    const backupPos = cur.POS || [0.5, 0.5];
                    cur.POS_S = cur.POS_S && cur.POS_S.length >= 2 ? cur.POS_S.slice(0, 2) : [backupPos, [backupPos[0]+0.05, backupPos[1]+0.05]];
                    if (typeof cur.RELEASE_MOUSE !== "boolean") cur.RELEASE_MOUSE = true;
                    if (typeof cur.SEPARAT !== "boolean") cur.SEPARAT = false;
                    if (typeof cur.TOUCH !== "boolean") cur.TOUCH = true;
                }));
            }
        }
    };
    React.useEffect(() => { ensureSmartToggleInit(); }, []);

    return (
        <Box sx={{ p: 1, mb: 1, bgcolor: '#fff', border: '1px solid #eee', borderRadius: 1 }}>
            <Grid container alignItems="center" spacing={1}>
                {isCombo ? (
                    <>
                        <Grid item xs={10}>
                            <Box sx={{ overflowX: 'auto', whiteSpace: 'nowrap' }}>
                                <Typography variant="body2" sx={{ fontWeight: 'bold', fontSize: '0.75rem', color: '#E91E63' }}>{data.KEY}</Typography>
                            </Box>
                        </Grid>
                        <Grid item xs={2}>
                            <IconButton size="small" onClick={() => setConfig(produce(d => { delete d.KEY_MAPS[data.KEY] }))}><HighlightOffIcon fontSize="small" /></IconButton>
                        </Grid>
                        <Grid item xs={12}>
                            <FormControl fullWidth size="small">
                                <Select value={data.TYPE} onChange={handleChange} sx={{ fontSize: '0.75rem', height: '28px' }}>
                                    <MenuItem value="PRESS">同步按住抬起</MenuItem>
                                    <MenuItem value="CLICK">单次点击</MenuItem>
                                    <MenuItem value="AUTO_FIRE">连发</MenuItem>
                                    <MenuItem value="DRAG">滑动</MenuItem>
                                    <MenuItem value="MULT_PRESS">多点触摸</MenuItem>
                                    <MenuItem value="SEQUENTIAL_PRESS">轮询点击</MenuItem>
                                    <MenuItem value="SYNC_VIEW_RESET">同步按抬重置视角</MenuItem>
                                    <MenuItem value="CLICK_VIEW_RESET">点击重置视角</MenuItem>
                                    <MenuItem value="BACKPACK_TOGGLE">背包键</MenuItem>
                                    <MenuItem value="SYNC_BACKPACK">同步按抬背包</MenuItem>
                                    <MenuItem value="SMART_TOGGLE">光标智能切换</MenuItem>
                                    <MenuItem value="PRESS_RELEASE_CLICK">按点抬点双击</MenuItem>
                                    <MenuItem value="RECOIL_SPEED_SET">压枪速度设置</MenuItem>
                                </Select>
                            </FormControl>
                        </Grid>
                    </>
                ) : (
                    <>
                        <Grid item xs={4}>
                            <Typography variant="body2" sx={{ fontWeight: 'bold', overflow: 'hidden', textOverflow: 'ellipsis', fontSize: '0.75rem' }}>{data.KEY}</Typography>
                        </Grid>
                        <Grid item xs={6}>
                            <FormControl fullWidth size="small">
                                <Select value={data.TYPE} onChange={handleChange} sx={{ fontSize: '0.75rem', height: '28px' }}>
                                    {!isWheel && <MenuItem value="PRESS">同步按住抬起</MenuItem>}
                                    <MenuItem value="CLICK">单次点击</MenuItem>
                                    {!isWheel && <MenuItem value="AUTO_FIRE">连发</MenuItem>}
                                    <MenuItem value="DRAG">滑动</MenuItem>
                                    <MenuItem value="MULT_PRESS">多点触摸</MenuItem>
                                    <MenuItem value="SEQUENTIAL_PRESS">轮询点击</MenuItem>
                                    {!isWheel && <MenuItem value="SYNC_VIEW_RESET">同步按抬重置视角</MenuItem>}
                                    {!isWheel && <MenuItem value="CLICK_VIEW_RESET">点击重置视角</MenuItem>}
                                    {!isWheel && <MenuItem value="BACKPACK_TOGGLE">背包键</MenuItem>}
                                    {!isWheel && <MenuItem value="SYNC_BACKPACK">同步按抬背包</MenuItem>}
                                    {!isWheel && <MenuItem value="SMART_TOGGLE">光标智能切换</MenuItem>}
                                    <MenuItem value="PRESS_RELEASE_CLICK">按点抬点双击</MenuItem>
                                    {!isWheel && <MenuItem value="RECOIL_SPEED_SET">压枪速度设置</MenuItem>}
                                </Select>
                            </FormControl>
                        </Grid>
                        <Grid item xs={2}>
                            <IconButton size="small" onClick={() => setConfig(produce(d => { delete d.KEY_MAPS[data.KEY] }))}><HighlightOffIcon fontSize="small" /></IconButton>
                        </Grid>
                    </>
                )}
            </Grid>

            <Box sx={{ mt: 1 }}>
                {["CLICK", "AUTO_FIRE", "SYNC_BACKPACK", "PRESS_RELEASE_CLICK", "BACKPACK_TOGGLE", "SEQUENTIAL_PRESS"].includes(data.TYPE) && (
                    <Grid container alignItems="center" sx={{ mb: 0.5 }}>
                        <Typography variant="caption" sx={{ minWidth: "50px" }}>点击时长:</Typography>
                        <CostumedInput 
                            defaultValue={data.TYPE === "AUTO_FIRE" ? data.INTERVAL?.[0] : (data.CLICK_DURATION || data.INTERVAL?.[0] || 18)} 
                            onCommit={v => setConfig(produce(d => { 
                                if(data.TYPE === "AUTO_FIRE") { if(!d.KEY_MAPS[data.KEY].INTERVAL) d.KEY_MAPS[data.KEY].INTERVAL = [18,20,10,10]; d.KEY_MAPS[data.KEY].INTERVAL[0] = Number(v); }
                                else if(data.TYPE === "CLICK") d.KEY_MAPS[data.KEY].INTERVAL = [Number(v)];
                                else d.KEY_MAPS[data.KEY].CLICK_DURATION = Number(v); 
                            }))} 
                            width="35px" 
                        />
                        <Typography variant="caption" sx={{ ml: 0.5 }}>ms</Typography>
                    </Grid>
                )}
                
                {data.TYPE === "AUTO_FIRE" && data.INTERVAL && (
                    <Box sx={{ mb: 0.5 }}>
                        <Grid container alignItems="center"><Typography variant="caption" sx={{ mr: 1 }}>间隔:</Typography><CostumedInput defaultValue={data.INTERVAL[1]} onCommit={v => setConfig(produce(d => { d.KEY_MAPS[data.KEY].INTERVAL[1] = Number(v) }))} width="35px" /><Typography variant="caption">ms</Typography></Grid>
                        <Grid container alignItems="center"><Typography variant="caption" sx={{ mr: 1 }}>随机时长:</Typography><CostumedInput defaultValue={data.INTERVAL[2]} onCommit={v => setConfig(produce(d => { d.KEY_MAPS[data.KEY].INTERVAL[2] = Number(v) }))} width="35px" /><Typography variant="caption">ms</Typography></Grid>
                        <Grid container alignItems="center"><Typography variant="caption" sx={{ mr: 1 }}>随机间隔:</Typography><CostumedInput defaultValue={data.INTERVAL[3]} onCommit={v => setConfig(produce(d => { d.KEY_MAPS[data.KEY].INTERVAL[3] = Number(v) }))} width="35px" /><Typography variant="caption">ms</Typography></Grid>
                    </Box>
                )}

                {data.TYPE === "SEQUENTIAL_PRESS" && (
                    <Grid container alignItems="center" sx={{ mb: 0.5 }}>
                        <Typography variant="caption" sx={{ minWidth: "50px" }}>冷却:</Typography>
                        <CostumedInput defaultValue={data.COOLDOWN || 0} onCommit={v => setConfig(produce(d => { d.KEY_MAPS[data.KEY].COOLDOWN = Number(v) }))} width="35px" />
                        <Typography variant="caption" sx={{ ml: 0.5 }}>ms</Typography>
                    </Grid>
                )}

                {["DRAG", "MULT_PRESS", "SEQUENTIAL_PRESS"].includes(data.TYPE) && (
                    <Box>
                        {data.TYPE === "DRAG" && <Grid container alignItems="center"><Typography variant="caption">间隔:</Typography><CostumedInput defaultValue={data.INTERVAL?.[0] || 18} onCommit={v => setConfig(produce(d => { d.KEY_MAPS[data.KEY].INTERVAL = [Number(v)] }))} width="35px" /> <Typography variant="caption">ms</Typography></Grid>}
                        {data.TYPE === "MULT_PRESS" && <Grid container alignItems="center"><Typography variant="caption">间隔:</Typography><CostumedInput defaultValue={data.INTERVAL?.[0] || 0} onCommit={v => setConfig(produce(d => { d.KEY_MAPS[data.KEY].INTERVAL = [Number(v)] }))} width="35px" /> <Typography variant="caption">ms</Typography></Grid>}
                        
                        <Button size="small" startIcon={<AddLocationIcon />} variant="outlined" sx={{ width: '100%', my: 0.5, fontSize: '0.65rem' }} 
                            onClick={() => onStartAddingPoint && onStartAddingPoint(data.KEY)}
                        >
                            {isAddingPoint && isAddingPoint(data.KEY) ? "请点击屏幕添加..." : "添加触点"}
                        </Button>
                        
                        <Box sx={{ maxHeight: 60, overflowY: 'auto', border: '1px solid #eee' }}>
                            {data.POS_S && data.POS_S.map((p, i) => (
                                <Grid container key={i} justifyContent="space-between" alignItems="center" sx={{ px: 0.5 }}>
                                    <Typography variant="caption">Pt{i+1}: ({parseInt(p[0]*100)},{parseInt(p[1]*100)})</Typography>
                                    <IconButton size="small" onClick={() => setConfig(produce(d => { d.KEY_MAPS[data.KEY].POS_S.splice(i, 1) }))}><HighlightOffIcon fontSize="inherit" /></IconButton>
                                </Grid>
                            ))}
                        </Box>
                    </Box>
                )}

                {["BACKPACK_TOGGLE", "SYNC_BACKPACK", "PRESS_RELEASE_CLICK"].includes(data.TYPE) && (
                    <Box>
                        <Grid container alignItems="center" justifyContent="space-between">
                            <Typography variant="caption">A: ({parseInt(data.POS?.[0]*100)},{parseInt(data.POS?.[1]*100)})</Typography>
                        </Grid>
                        <Grid container alignItems="center" justifyContent="space-between">
                            <Typography variant="caption">B: ({parseInt(data.POS_B?.[0]*100)},{parseInt(data.POS_B?.[1]*100)})</Typography>
                            <Button size="small" startIcon={<RestartAltIcon />} variant="text" sx={{ fontSize: '0.65rem', p: 0, minWidth: 'auto' }} 
                                onClick={() => onStartSettingPosB && onStartSettingPosB(data.KEY)}
                            >
                                {isSettingPosB && isSettingPosB(data.KEY) ? "点击屏幕设置..." : "重设"}
                            </Button>
                        </Grid>
                    </Box>
                )}

                {data.TYPE === "SMART_TOGGLE" && (
                    <Box sx={{ mt: 1, p: 0.5, bgcolor: '#f3e5f5', borderRadius: 1 }}>
                        <Grid container spacing={1}>
                            <Grid item xs={12} sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                <Typography variant="caption" sx={{ fontWeight: 'bold', color: '#9C27B0' }}>
                                    触点A (去): ({parseInt(data.POS_S?.[0]?.[0]*100)},{parseInt(data.POS_S?.[0]?.[1]*100)})
                                </Typography>
                                <Button size="small" variant="outlined" sx={{ fontSize: '0.6rem', p: 0 }} onClick={() => onStartSettingSmartToggleIndex && onStartSettingSmartToggleIndex({key: data.KEY, index: 0})}>
                                    {isSettingSmartToggleIndex && isSettingSmartToggleIndex(data.KEY, 0) ? "点击屏幕..." : "重设A"}
                                </Button>
                            </Grid>
                            <Grid item xs={12} sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                <Typography variant="caption" sx={{ fontWeight: 'bold', color: '#9C27B0' }}>
                                    触点B (回): ({parseInt(data.POS_S?.[1]?.[0]*100)},{parseInt(data.POS_S?.[1]?.[1]*100)})
                                </Typography>
                                <Button size="small" variant="outlined" sx={{ fontSize: '0.6rem', p: 0 }} onClick={() => onStartSettingSmartToggleIndex && onStartSettingSmartToggleIndex({key: data.KEY, index: 1})}>
                                    {isSettingSmartToggleIndex && isSettingSmartToggleIndex(data.KEY, 1) ? "点击屏幕..." : "重设B"}
                                </Button>
                            </Grid>
                            <Grid item xs={12}>
                                <Grid container alignItems="center" justifyContent="space-between">
                                    <Typography variant="caption">轮询:第一次按点A，二次按点B</Typography>
                                    <Switch size="small" checked={data.SEPARAT} onChange={() => setConfig(produce(d => { d.KEY_MAPS[data.KEY].SEPARAT = !d.KEY_MAPS[data.KEY].SEPARAT }))} />
                                </Grid>
                            </Grid>
                            <Grid item xs={12}>
                                <Grid container alignItems="center" justifyContent="space-between">
                                    <Typography variant="caption">触发时释放光标</Typography>
                                    <Switch size="small" checked={data.RELEASE_MOUSE} onChange={() => setConfig(produce(d => { d.KEY_MAPS[data.KEY].RELEASE_MOUSE = !d.KEY_MAPS[data.KEY].RELEASE_MOUSE }))} />
                                </Grid>
                            </Grid>
                            <Grid item xs={12}>
                                <Grid container alignItems="center" justifyContent="space-between">
                                    <Typography variant="caption">执行屏幕触摸点击</Typography>
                                    <Switch size="small" checked={data.TOUCH} onChange={() => setConfig(produce(d => { d.KEY_MAPS[data.KEY].TOUCH = !d.KEY_MAPS[data.KEY].TOUCH }))} />
                                </Grid>
                            </Grid>
                        </Grid>
                    </Box>
                )}
                
                {data.TYPE === "RECOIL_SPEED_SET" && (
                    <Box>
                        <Grid container alignItems="center"><Typography variant="caption">速度:</Typography><CostumedInput defaultValue={data.VALUE || 0} onCommit={v => setConfig(produce(d => { d.KEY_MAPS[data.KEY].VALUE = Number(v) }))} width="40px" /></Grid>
                        <Grid container alignItems="center"><Typography variant="caption">备注:</Typography><CostumedStringInput defaultValue={data.NOTE || ""} onCommit={v => setConfig(produce(d => { d.KEY_MAPS[data.KEY].NOTE = v }))} width="100px" /></Grid>
                    </Box>
                )}
            </Box>
        </Box>
    );
};

const KeysSettings = ({ config, setConfig, setSelectKEY, visualSwitches, setVisualSwitches, onStartAddingPoint, isAddingPoint, onStartSettingPosB, isSettingPosB, onStartSettingSmartToggleIndex, isSettingSmartToggleIndex, onMouseBtnClick }) => {
    const singleKeys = Object.keys(config.KEY_MAPS).filter(k => !k.includes("+"));

    return (
        <Box sx={{ p: 1 }}>
            <Typography variant="subtitle2" sx={{color: "#9C27B0", mb: 1, fontWeight: 'bold', fontSize: '0.8rem'}}>按键配置</Typography>
            
            <Grid container alignItems="center" sx={{ mb: 1, bgcolor: '#f3e5f5', p: 0.5, borderRadius: 1 }}>
                <Typography variant="caption" sx={{ mr: 1 }}>按键位置可视化:</Typography>
                <Switch size="small" checked={visualSwitches.keys} onChange={() => setVisualSwitches(produce(d => { d.keys = !d.keys }))} />
            </Grid>

            <Typography variant="caption" sx={{color: '#666', fontSize: '0.7rem'}}>快捷添加鼠标键:</Typography>
            <Grid container spacing={0.5} sx={{ mb: 1 }}>
                {["BTN_LEFT", "BTN_RIGHT", "BTN_MIDDLE", "BTN_SIDE", "BTN_EXTRA", "REL_WHEEL_UP", "REL_WHEEL_DOWN"].map(btn => (
                    <Grid item key={btn} xs={4}>
                        <Button variant="outlined" size="small" fullWidth sx={{fontSize: '0.65rem', p:0}} 
                            onClick={() => {
                                setSelectKEY(btn); 
                                if (onMouseBtnClick) onMouseBtnClick(btn); 
                            }}
                        >
                            {btn.replace("REL_", "").replace("BTN_", "")}
                        </Button>
                    </Grid>
                ))}
            </Grid>
            <Box sx={{ border: '1px solid #ddd', borderRadius: 1, p: 0.5, mb: 1 }}>
                <Grid container alignItems="center">
                    <Typography variant="body2" sx={{ mr: 1, fontSize: '0.75rem' }}>随机落点（抖动）:</Typography>
                    <Switch size="small" checked={config.KEY_JITTER.ENABLE} onChange={() => setConfig(produce(d => { d.KEY_JITTER.ENABLE = !d.KEY_JITTER.ENABLE }))} />
                </Grid>
                <SliderWithInput 
                    label="范围幅度%" 
                    disabled={!config.KEY_JITTER.ENABLE} 
                    value={config.KEY_JITTER.AMOUNT * 100} 
                    min={0} 
                    max={5} 
                    step={0.05} 
                    onChange={v => setConfig(produce(d => { d.KEY_JITTER.AMOUNT = Number(v) / 100 }))} 
                />
            </Box>
            <Typography variant="subtitle2" sx={{ mb: 0.5, mt: 1, fontSize: '0.8rem' }}>按键列表:</Typography>
            {singleKeys.map(key => (
                <KeyMapSettings 
                    key={key} 
                    data={{ ...config.KEY_MAPS[key], KEY: key }} 
                    config={config} 
                    setConfig={setConfig} 
                    onStartAddingPoint={onStartAddingPoint}
                    isAddingPoint={isAddingPoint}
                    onStartSettingPosB={onStartSettingPosB}
                    isSettingPosB={isSettingPosB}
                    onStartSettingSmartToggleIndex={onStartSettingSmartToggleIndex}
                    isSettingSmartToggleIndex={isSettingSmartToggleIndex}
                />
            ))}
        </Box>
    );
};

const ComboSettings = ({ config, setConfig, visualSwitches, setVisualSwitches, onStartAddingPoint, isAddingPoint, onStartSettingPosB, isSettingPosB, onStartSettingSmartToggleIndex, isSettingSmartToggleIndex }) => {
    const comboKeys = Object.keys(config.KEY_MAPS).filter(k => k.includes("+"));

    return (
        <Box sx={{ p: 1 }}>
            <Typography variant="subtitle2" sx={{color: "#E91E63", mb: 1, fontWeight: 'bold', fontSize: '0.8rem'}}>组合键配置</Typography>
            
            <Grid container alignItems="center" sx={{ mb: 1, bgcolor: '#fce4ec', p: 0.5, borderRadius: 1 }}>
                <Typography variant="caption" sx={{ mr: 1 }}>胶囊位置可视化:</Typography>
                <Switch size="small" checked={visualSwitches.combos} onChange={() => setVisualSwitches(produce(d => { d.combos = !d.combos }))} />
            </Grid>

            <Box sx={{ mb: 2, p: 1, bgcolor: '#eee', borderRadius: 1 }}>
                <Typography variant="caption" sx={{ display: 'block', color: '#666' }}>提示: 按住键 再按另一个键，点击背景以添加组合键 </Typography>
                <Typography variant="caption" sx={{ display: 'block', color: '#666' }}>显示鼠标键时按住其他键，可以与其他键组合 </Typography>
            </Box>

            <Typography variant="subtitle2" sx={{ mb: 0.5, mt: 1, fontSize: '0.8rem' }}>已添加的组合键列表:</Typography>
            {comboKeys.map(key => (
                <KeyMapSettings 
                    key={key} 
                    data={{ ...config.KEY_MAPS[key], KEY: key }} 
                    config={config} 
                    setConfig={setConfig} 
                    isCombo={true}
                    onStartAddingPoint={onStartAddingPoint}
                    isAddingPoint={isAddingPoint}
                    onStartSettingPosB={onStartSettingPosB}
                    isSettingPosB={isSettingPosB}
                    onStartSettingSmartToggleIndex={onStartSettingSmartToggleIndex}
                    isSettingSmartToggleIndex={isSettingSmartToggleIndex}
                />
            ))}
        </Box>
    );
};

export { KeysSettings, ComboSettings };