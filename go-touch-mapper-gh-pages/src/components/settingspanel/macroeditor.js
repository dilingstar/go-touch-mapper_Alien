import React, { useState } from 'react';
import { Grid, Button, Typography, Switch, Box, IconButton, Select, MenuItem, FormControl, TextField, Menu } from "@mui/material";
import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';
import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import CloseIcon from '@mui/icons-material/Close'; // [V3.4.3] 替换删除图标
import { produce } from "immer";
import { CostumedInput } from "../UIcomponents";

// 子组件：事件编辑器
const EventItem = ({ event, index, onDelete, onMoveUp, onMoveDown, onChange, onAddPoint, isAddingPoint }) => {
    // 渲染事件名称
    const getEventName = (type) => {
        const map = { "CLICK": "点击", "SWIPE": "滑动", "DELAY": "延迟", "MAP_ON": "开启映射", "MAP_OFF": "关闭映射" };
        return map[type] || type;
    };

    return (
        <Box sx={{ p: 0.5, mb: 0.5, border: '1px solid #eee', borderRadius: 1, bgcolor: '#fff' }}>
            <Grid container alignItems="center" spacing={0.5}>
                <Grid item xs={1}><Typography variant="caption" sx={{color:'#999'}}>{index+1}</Typography></Grid>
                <Grid item xs={3}>
                    <Typography variant="body2" sx={{fontSize: '0.7rem', fontWeight: 'bold'}}>{getEventName(event.TYPE)}</Typography>
                </Grid>
                
                <Grid item xs={6}>
                    {event.TYPE === "CLICK" && (
                        <Grid container alignItems="center">
                            <CostumedInput defaultValue={event.DURATION || 50} onCommit={v => onChange(produce(event, d => { d.DURATION = Number(v) }))} width="30px" />
                            <Typography variant="caption" sx={{ml:0.5, mr:0.5}}>ms</Typography>
                            <Button size="small" variant={isAddingPoint ? "contained" : "outlined"} sx={{ fontSize: '0.6rem', minWidth: '30px', p:0 }} onClick={onAddPoint}>
                                {isAddingPoint ? "取点中" : "位置"}
                            </Button>
                        </Grid>
                    )}
                    {event.TYPE === "SWIPE" && (
                        <Grid container alignItems="center">
                            <CostumedInput defaultValue={event.DURATION || 20} onCommit={v => onChange(produce(event, d => { d.DURATION = Number(v) }))} width="30px" />
                            <Typography variant="caption" sx={{ml:0.5, mr:0.5}}>ms</Typography>
                            <Button size="small" variant={isAddingPoint ? "contained" : "outlined"} sx={{ fontSize: '0.6rem', minWidth: '30px', p:0 }} onClick={onAddPoint}>
                                {isAddingPoint ? "取点中" : "路径+"}
                            </Button>
                        </Grid>
                    )}
                    {event.TYPE === "DELAY" && (
                        <Grid container alignItems="center">
                            <CostumedInput defaultValue={event.DURATION || 100} onCommit={v => onChange(produce(event, d => { d.DURATION = Number(v) }))} width="40px" />
                            <Typography variant="caption" sx={{ml:0.5}}>ms</Typography>
                        </Grid>
                    )}
                </Grid>

                <Grid item xs={2} sx={{ display: 'flex', justifyContent: 'flex-end' }}>
                    <IconButton size="small" onClick={onMoveUp} disabled={index === 0} sx={{ p: 0 }}><KeyboardArrowUpIcon fontSize="small" /></IconButton>
                    <IconButton size="small" onClick={onMoveDown} sx={{ p: 0 }}><KeyboardArrowDownIcon fontSize="small" /></IconButton>
                    {/* [V3.4.3] 替换为 CloseIcon */}
                    <IconButton size="small" onClick={onDelete} sx={{ p: 0, color: '#f44336' }}><CloseIcon fontSize="small" /></IconButton>
                </Grid>
            </Grid>
            
            {/* 滑动路径点列表 */}
            {event.TYPE === "SWIPE" && event.POS_LIST && event.POS_LIST.length > 0 && (
                <Box sx={{ pl: 2, mt: 0.5 }}>
                    {event.POS_LIST.map((p, pIdx) => (
                        <Typography key={pIdx} variant="caption" sx={{ display: 'block', color: '#666' }}>
                            P{pIdx+1}: ({parseInt(p[0]*100)},{parseInt(p[1]*100)})
                            <span style={{color:'red', cursor:'pointer', marginLeft:'5px'}} onClick={() => onChange(produce(event, d => { d.POS_LIST.splice(pIdx, 1) }))}>[x]</span>
                        </Typography>
                    ))}
                </Box>
            )}
        </Box>
    );
};

// 子组件：宏卡片
const MacroCard = ({ macro, index, config, setConfig, currentActiveKey, addingMacroPoint, setAddingMacroPoint }) => {
    const updateMacro = (fn) => setConfig(produce(d => { if(d.MACROS[index]) fn(d.MACROS[index]) }));
    const deleteMacro = () => setConfig(produce(d => { d.MACROS.splice(index, 1) }));

    const [anchorEl, setAnchorEl] = useState(null);
    const [targetListField, setTargetListField] = useState(null);

    const handleAddClick = (event, field) => {
        setAnchorEl(event.currentTarget);
        setTargetListField(field);
    };

    const handleAddClose = (type) => {
        if (type && targetListField) {
            updateMacro(m => {
                if (!m[targetListField]) m[targetListField] = [];
                let newEvent = { TYPE: type, DURATION: 50 };
                if (type === "CLICK") newEvent.POS = [0.5, 0.5];
                if (type === "SWIPE") newEvent.POS_LIST = [];
                if (type === "DELAY") newEvent.DURATION = 100;
                m[targetListField].push(newEvent);
            });
        }
        setAnchorEl(null);
        setTargetListField(null);
    };

    const bindKey = (targetField) => {
        if (currentActiveKey) {
            updateMacro(m => { m[targetField] = currentActiveKey });
        }
    };

    let borderColor = '#ccc';
    if (macro.TYPE === "PRESS_RELEASE") borderColor = '#2196F3';
    if (macro.TYPE === "HOLD_LOOP") borderColor = '#FF9800';
    if (macro.TYPE === "TOGGLE_LOOP") borderColor = '#4CAF50';

    return (
        <Box sx={{ mb: 2, border: `2px solid ${borderColor}`, borderRadius: 2, p: 1, bgcolor: '#fafafa' }}>
            <Grid container alignItems="center" justifyContent="space-between" sx={{ mb: 1 }}>
                <TextField 
                    variant="standard" 
                    value={macro.LABEL} 
                    onChange={(e) => updateMacro(m => { m.LABEL = e.target.value })}
                    inputProps={{ style: { fontSize: '0.8rem', fontWeight: 'bold' } }}
                />
                <IconButton size="small" onClick={deleteMacro} sx={{ color: '#f44336' }}><DeleteIcon /></IconButton>
            </Grid>

            {/* 触发键绑定 */}
            <Grid container alignItems="center" spacing={1} sx={{ mb: 1 }}>
                <Grid item xs={3}><Typography variant="caption">触发键:</Typography></Grid>
                <Grid item xs={6}>
                    <Box sx={{ border: '1px solid #ddd', borderRadius: 1, px: 1, py: 0.5, minHeight: '24px', display: 'flex', alignItems: 'center', bgcolor: '#fff' }}>
                        <Typography variant="caption" sx={{ fontWeight: 'bold', color: '#1976D2' }}>{macro.TRIGGER_KEY || "未绑定"}</Typography>
                        {macro.TRIGGER_KEY && <span style={{marginLeft:'auto', cursor:'pointer', color:'#999'}} onClick={() => updateMacro(m => m.TRIGGER_KEY = "")}>×</span>}
                    </Box>
                </Grid>
                <Grid item xs={3}>
                    <Button 
                        size="small" variant={currentActiveKey ? "contained" : "outlined"} 
                        disabled={!currentActiveKey}
                        sx={{ fontSize: '0.6rem', p: 0, minWidth: '40px' }} 
                        onClick={() => bindKey("TRIGGER_KEY")}
                    >
                        绑定
                    </Button>
                </Grid>
            </Grid>

            {/* 终止键 */}
            {macro.TYPE === "TOGGLE_LOOP" && (
                <Grid container alignItems="center" spacing={1} sx={{ mb: 1 }}>
                    <Grid item xs={3}><Typography variant="caption">终止键:</Typography></Grid>
                    <Grid item xs={6}>
                        <Box sx={{ border: '1px solid #ddd', borderRadius: 1, px: 1, py: 0.5, minHeight: '24px', display: 'flex', alignItems: 'center', bgcolor: '#fff' }}>
                            <Typography variant="caption" sx={{ fontWeight: 'bold', color: '#E64A19' }}>{macro.STOP_KEY || "无"}</Typography>
                            {macro.STOP_KEY && <span style={{marginLeft:'auto', cursor:'pointer', color:'#999'}} onClick={() => updateMacro(m => m.STOP_KEY = "")}>×</span>}
                        </Box>
                    </Grid>
                    <Grid item xs={3}>
                        <Button 
                            size="small" variant={currentActiveKey ? "contained" : "outlined"} 
                            disabled={!currentActiveKey}
                            sx={{ fontSize: '0.6rem', p: 0, minWidth: '40px' }} 
                            onClick={() => bindKey("STOP_KEY")}
                        >
                            绑定
                        </Button>
                    </Grid>
                </Grid>
            )}

            {/* 触发模式 */}
            <Grid container alignItems="center" spacing={1} sx={{ mb: 1 }}>
                <Grid item xs={4}><Typography variant="caption">状态触发:</Typography></Grid>
                <Grid item xs={8}>
                    <FormControl fullWidth size="small">
                        <Select 
                            value={macro.TRIGGER_MODE} 
                            onChange={(e) => updateMacro(m => m.TRIGGER_MODE = e.target.value)}
                            sx={{ fontSize: '0.7rem', height: '24px' }}
                        >
                            <MenuItem value="ALWAYS">都触发</MenuItem>
                            <MenuItem value="MAP_ON">开启映射时</MenuItem>
                            <MenuItem value="MAP_OFF">关闭映射时</MenuItem>
                        </Select>
                    </FormControl>
                </Grid>
            </Grid>

            {/* 循环间隔 */}
            {(macro.TYPE === "HOLD_LOOP" || macro.TYPE === "TOGGLE_LOOP") && (
                <Grid container alignItems="center" spacing={1} sx={{ mb: 1 }}>
                    <Grid item xs={4}><Typography variant="caption">循环间隔:</Typography></Grid>
                    <Grid item xs={4}><CostumedInput defaultValue={macro.LOOP_INTERVAL} onCommit={v => updateMacro(m => m.LOOP_INTERVAL = Number(v))} width="50px" /></Grid>
                    <Grid item xs={4}><Typography variant="caption">ms</Typography></Grid>
                </Grid>
            )}

            {/* 按下事件列表 */}
            <Box sx={{ mt: 1, borderTop: '1px dashed #ccc', pt: 1 }}>
                <Grid container justifyContent="space-between" alignItems="center">
                    <Typography variant="caption" sx={{ fontWeight: 'bold' }}>
                        {macro.TYPE === "PRESS_RELEASE" ? "按下执行:" : "循环执行列表:"}
                    </Typography>
                    <Button size="small" startIcon={<AddIcon />} sx={{ fontSize: '0.6rem' }} onClick={(e) => handleAddClick(e, "PRESS_EVENTS")}>添加</Button>
                </Grid>
                
                <Grid container alignItems="center" spacing={1} sx={{ mb: 0.5 }}>
                    <Grid item xs={3}><Typography variant="caption">模式:</Typography></Grid>
                    <Grid item xs={9}>
                        <FormControl fullWidth size="small">
                            <Select 
                                value={macro.EXEC_MODE_PRESS} 
                                onChange={(e) => updateMacro(m => m.EXEC_MODE_PRESS = e.target.value)}
                                sx={{ fontSize: '0.7rem', height: '24px' }}
                            >
                                <MenuItem value="SEQUENCE">依次触发</MenuItem>
                                <MenuItem value="TIMEOUT">超时触发</MenuItem>
                                <MenuItem value="SIMULTANEOUS">同时触发</MenuItem>
                            </Select>
                        </FormControl>
                    </Grid>
                </Grid>
                {macro.EXEC_MODE_PRESS === "TIMEOUT" && (
                    <Grid container alignItems="center" sx={{ mb: 0.5, pl: 2 }}>
                        <Typography variant="caption" sx={{ mr: 1 }}>限时:</Typography>
                        <CostumedInput defaultValue={macro.TIMEOUT_PRESS} onCommit={v => updateMacro(m => m.TIMEOUT_PRESS = Number(v))} width="40px" />
                        <Typography variant="caption" sx={{ ml: 0.5 }}>ms</Typography>
                    </Grid>
                )}

                {macro.PRESS_EVENTS && macro.PRESS_EVENTS.map((evt, idx) => (
                    <EventItem 
                        key={idx} event={evt} index={idx} 
                        onChange={(newEvt) => updateMacro(m => m.PRESS_EVENTS[idx] = newEvt)}
                        onDelete={() => updateMacro(m => m.PRESS_EVENTS.splice(idx, 1))}
                        onMoveUp={() => updateMacro(m => { if(idx>0) { const tmp=m.PRESS_EVENTS[idx]; m.PRESS_EVENTS[idx]=m.PRESS_EVENTS[idx-1]; m.PRESS_EVENTS[idx-1]=tmp; } })}
                        onMoveDown={() => updateMacro(m => { if(idx<m.PRESS_EVENTS.length-1) { const tmp=m.PRESS_EVENTS[idx]; m.PRESS_EVENTS[idx]=m.PRESS_EVENTS[idx+1]; m.PRESS_EVENTS[idx+1]=tmp; } })}
                        onAddPoint={() => setAddingMacroPoint({ macroId: macro.ID, eventType: "PRESS", eventIndex: idx, isSwipe: evt.TYPE === "SWIPE" })}
                        isAddingPoint={addingMacroPoint && addingMacroPoint.macroId === macro.ID && addingMacroPoint.eventIndex === idx && addingMacroPoint.eventType === "PRESS"}
                    />
                ))}
            </Box>

            {/* 抬起事件列表 */}
            {macro.TYPE === "PRESS_RELEASE" && (
                <Box sx={{ mt: 1, borderTop: '1px dashed #ccc', pt: 1 }}>
                    <Grid container justifyContent="space-between" alignItems="center">
                        <Typography variant="caption" sx={{ fontWeight: 'bold' }}>抬起执行:</Typography>
                        <Button size="small" startIcon={<AddIcon />} sx={{ fontSize: '0.6rem' }} onClick={(e) => handleAddClick(e, "RELEASE_EVENTS")}>添加</Button>
                    </Grid>
                    
                    <Grid container alignItems="center" spacing={1} sx={{ mb: 0.5 }}>
                        <Grid item xs={3}><Typography variant="caption">模式:</Typography></Grid>
                        <Grid item xs={9}>
                            <FormControl fullWidth size="small">
                                <Select 
                                    value={macro.EXEC_MODE_RELEASE} 
                                    onChange={(e) => updateMacro(m => m.EXEC_MODE_RELEASE = e.target.value)}
                                    sx={{ fontSize: '0.7rem', height: '24px' }}
                                >
                                    <MenuItem value="SEQUENCE">依次触发</MenuItem>
                                    <MenuItem value="TIMEOUT">超时触发</MenuItem>
                                    <MenuItem value="SIMULTANEOUS">同时触发</MenuItem>
                                </Select>
                            </FormControl>
                        </Grid>
                    </Grid>
                    {macro.EXEC_MODE_RELEASE === "TIMEOUT" && (
                        <Grid container alignItems="center" sx={{ mb: 0.5, pl: 2 }}>
                            <Typography variant="caption" sx={{ mr: 1 }}>限时:</Typography>
                            <CostumedInput defaultValue={macro.TIMEOUT_RELEASE} onCommit={v => updateMacro(m => m.TIMEOUT_RELEASE = Number(v))} width="40px" />
                            <Typography variant="caption" sx={{ ml: 0.5 }}>ms</Typography>
                        </Grid>
                    )}

                    {macro.RELEASE_EVENTS && macro.RELEASE_EVENTS.map((evt, idx) => (
                        <EventItem 
                            key={idx} event={evt} index={idx} 
                            onChange={(newEvt) => updateMacro(m => m.RELEASE_EVENTS[idx] = newEvt)}
                            onDelete={() => updateMacro(m => m.RELEASE_EVENTS.splice(idx, 1))}
                            onMoveUp={() => updateMacro(m => { if(idx>0) { const tmp=m.RELEASE_EVENTS[idx]; m.RELEASE_EVENTS[idx]=m.RELEASE_EVENTS[idx-1]; m.RELEASE_EVENTS[idx-1]=tmp; } })}
                            onMoveDown={() => updateMacro(m => { if(idx<m.RELEASE_EVENTS.length-1) { const tmp=m.RELEASE_EVENTS[idx]; m.RELEASE_EVENTS[idx]=m.RELEASE_EVENTS[idx+1]; m.RELEASE_EVENTS[idx+1]=tmp; } })}
                            onAddPoint={() => setAddingMacroPoint({ macroId: macro.ID, eventType: "RELEASE", eventIndex: idx, isSwipe: evt.TYPE === "SWIPE" })}
                            isAddingPoint={addingMacroPoint && addingMacroPoint.macroId === macro.ID && addingMacroPoint.eventIndex === idx && addingMacroPoint.eventType === "RELEASE"}
                        />
                    ))}
                </Box>
            )}

            <Menu
                anchorEl={anchorEl}
                open={Boolean(anchorEl)}
                onClose={() => handleAddClose(null)}
            >
                <MenuItem onClick={() => handleAddClose("CLICK")}>点击</MenuItem>
                <MenuItem onClick={() => handleAddClose("SWIPE")}>滑动</MenuItem>
                <MenuItem onClick={() => handleAddClose("DELAY")}>延迟</MenuItem>
                <MenuItem onClick={() => handleAddClose("MAP_ON")}>开启映射</MenuItem>
                <MenuItem onClick={() => handleAddClose("MAP_OFF")}>关闭映射</MenuItem>
            </Menu>
        </Box>
    );
};

export default function MacroEditor({ config, setConfig, currentActiveKey, visualSwitches, setVisualSwitches, addingMacroPoint, setAddingMacroPoint }) {
    const [createAnchorEl, setCreateAnchorEl] = useState(null);

    const addMacro = (type, defaultLabel) => {
        const newId = `macro_${Date.now()}`;
        const count = config.MACROS ? config.MACROS.filter(m => m.TYPE === type).length + 1 : 1;
        const label = `${defaultLabel}${count}`;

        const newMacro = {
            ID: newId,
            LABEL: label,
            TRIGGER_KEY: "",
            TRIGGER_MODE: "ALWAYS", 
            TYPE: type, 
            LOOP_INTERVAL: 50,
            PRESS_EVENTS: [],
            RELEASE_EVENTS: [],
            EXEC_MODE_PRESS: "SEQUENCE",
            EXEC_MODE_RELEASE: "SEQUENCE",
            TIMEOUT_PRESS: 50,
            TIMEOUT_RELEASE: 50
        };

        setConfig(produce(d => {
            if (!d.MACROS) d.MACROS = [];
            d.MACROS.push(newMacro);
        }));
        setCreateAnchorEl(null);
    };

    return (
        <div style={{ padding: '8px' }}>
            <Typography variant="subtitle2" sx={{color: "#607D8B", mb: 1, fontWeight: 'bold', fontSize: '0.8rem'}}>宏编辑器配置 v1</Typography>
            
            <Grid container alignItems="center" sx={{ mb: 2, bgcolor: '#eceff1', p: 0.5, borderRadius: 1 }}>
                <Typography variant="caption" sx={{ mr: 1 }}>宏执行可视化:</Typography>
                <Switch size="small" checked={visualSwitches.macros} onChange={() => setVisualSwitches(produce(d => { d.macros = !d.macros }))} />
            </Grid>

            <Button 
                variant="contained" fullWidth size="small" sx={{ mb: 2 }}
                onClick={(e) => setCreateAnchorEl(e.currentTarget)}
            >
                创建一个新宏模板
            </Button>
            
            <Menu
                anchorEl={createAnchorEl}
                open={Boolean(createAnchorEl)}
                onClose={() => setCreateAnchorEl(null)}
            >
                <MenuItem onClick={() => addMacro("PRESS_RELEASE", "按抬模板")}>按抬模板</MenuItem>
                <MenuItem onClick={() => addMacro("HOLD_LOOP", "按住循环模板")}>按住循环模板</MenuItem>
                <MenuItem onClick={() => addMacro("TOGGLE_LOOP", "开关循环模板")}>开关循环模板</MenuItem>
            </Menu>

            <Typography variant="body2" sx={{ mb: 1, fontSize: '0.75rem', fontWeight: 'bold' }}>已添加的宏模板列表:</Typography>
            
            {(!config.MACROS || config.MACROS.length === 0) && (
                <Typography variant="caption" sx={{ color: '#999', display: 'block', textAlign: 'center', mt: 2 }}>暂无宏，请点击上方按钮创建</Typography>
            )}

            {config.MACROS && config.MACROS.map((macro, index) => (
                <MacroCard 
                    key={macro.ID} 
                    index={index} 
                    macro={macro} 
                    config={config} 
                    setConfig={setConfig} 
                    currentActiveKey={currentActiveKey}
                    addingMacroPoint={addingMacroPoint}
                    setAddingMacroPoint={setAddingMacroPoint}
                />
            ))}
        </div>
    );
}