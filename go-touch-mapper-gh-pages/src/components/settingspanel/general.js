import React, { useState, useRef, useEffect } from 'react';
import { Grid, Button, Typography, Box, Badge, Switch, Divider } from "@mui/material";
import InfoIcon from '@mui/icons-material/Info';
import HighlightOffIcon from '@mui/icons-material/HighlightOff';
import { produce } from "immer";
import { CostumedInput } from "../UIcomponents";
import InfoModal from "../InfoModal";

// Version: V3.5.0

export default function GeneralSettings({ 
    config, setConfig, 
    getRemoteApiImg, openFileInput, 
    continuousShot, toggleContinuousShot, 
    intervalMs, setIntervalMs,
    infoBadgeVisible, setInfoBadgeVisible,
    currentActiveKey // [V3.5.0] 接收传进来的按键状态，用于快捷绑定
}) {
    const [showInfoModal, setShowInfoModal] = useState(false);

    const handleOpenInfo = () => {
        setShowInfoModal(true);
        if (setInfoBadgeVisible) {
            setInfoBadgeVisible(false);
        }
    };

    // --- 快捷结束程序逻辑 ---
    // exitBtnState: 0=初始橙黄, 1=确认倒计时(淡红色), 2=展示消息状态
    const [exitBtnState, setExitBtnState] = useState(0); 
    const [countdown, setCountdown] = useState(5);
    const [exitMessage, setExitMessage] = useState("");
    const timerRef = useRef(null);

    useEffect(() => {
        return () => {
            if (timerRef.current) clearInterval(timerRef.current);
        };
    }, []);

    const handleExitClick = () => {
        if (exitBtnState === 0) {
            setExitBtnState(1);
            setCountdown(5);
            if (timerRef.current) clearInterval(timerRef.current);
            timerRef.current = setInterval(() => {
                setCountdown((prev) => {
                    if (prev <= 1) {
                        clearInterval(timerRef.current);
                        setExitBtnState(0);
                        return 5;
                    }
                    return prev - 1;
                });
            }, 1000);
        } else if (exitBtnState === 1) {
            clearInterval(timerRef.current);
            fetch('/api/exit', { method: 'POST' })
                .then(res => res.text())
                .then(text => {
                    setExitMessage(text);
                    setExitBtnState(2);
                    setTimeout(() => {
                        setExitMessage("");
                        setExitBtnState(0);
                    }, 2000);
                })
                .catch(err => {
                    setExitMessage("连接失败");
                    setExitBtnState(2);
                    setTimeout(() => {
                        setExitMessage("");
                        setExitBtnState(0);
                    }, 2000);
                });
        }
    };

    // --- 快捷绑定配置操作 ---
    const isExitEnable = config.END_EXIT?.EXIT_ENABLE || false;
    const exitKeys = config.END_EXIT?.EXIT_KEYS || [];

    const toggleExitEnable = () => {
        setConfig(produce(d => {
            if (!d.END_EXIT) d.END_EXIT = { EXIT_ENABLE: false, EXIT_KEYS: [] };
            d.END_EXIT.EXIT_ENABLE = !d.END_EXIT.EXIT_ENABLE;
        }));
    };

    const addExitKey = () => {
        if (currentActiveKey) {
            setConfig(produce(d => {
                if (!d.END_EXIT) d.END_EXIT = { EXIT_ENABLE: false, EXIT_KEYS: [] };
                if (!d.END_EXIT.EXIT_KEYS) d.END_EXIT.EXIT_KEYS = [];
                if (d.END_EXIT.EXIT_KEYS.indexOf(currentActiveKey) === -1) {
                    d.END_EXIT.EXIT_KEYS.push(currentActiveKey);
                }
            }));
        }
    };

    const removeExitKey = (idx) => {
        setConfig(produce(d => {
            d.END_EXIT.EXIT_KEYS.splice(idx, 1);
        }));
    };

    return (
        <Box sx={{ p: 1 }}>
            <Typography variant="subtitle2" sx={{color: "#607D8B", mb: 1, fontWeight: 'bold', fontSize: '0.8rem'}}>常规设置</Typography>
            <Grid container alignItems="center" spacing={1} sx={{ mb: 2 }}>
                <Grid item><Typography variant="body2" sx={{fontSize: '0.75rem'}}>分辨率:</Typography></Grid>
                <Grid item><CostumedInput defaultValue={config.SCREEN.SIZE[0]} onCommit={v => setConfig(produce(d => { d.SCREEN.SIZE[0] = Number(v) }))} width="40px" /></Grid>
                <Grid item><Typography variant="body2" sx={{fontSize: '0.75rem'}}>x</Typography></Grid>
                <Grid item><CostumedInput defaultValue={config.SCREEN.SIZE[1]} onCommit={v => setConfig(produce(d => { d.SCREEN.SIZE[1] = Number(v) }))} width="40px" /></Grid>
            </Grid>
            <Grid container spacing={1} sx={{ mb: 2 }}>
                <Grid item xs={6}><Button variant="outlined" fullWidth size="small" sx={{fontSize: '0.7rem'}} onClick={() => getRemoteApiImg("/screen.png?t=" + Date.now())}>获取截图</Button></Grid>
                <Grid item xs={6}><Button variant="outlined" fullWidth size="small" sx={{fontSize: '0.7rem'}} onClick={() => setTimeout(() => getRemoteApiImg("/screen.png?t=" + Date.now()), 5000)}>5秒后截图</Button></Grid>
                <Grid item xs={12}><Button variant="outlined" fullWidth size="small" sx={{fontSize: '0.7rem'}} onClick={openFileInput}>上传图片</Button></Grid>
                
                <Grid item xs={7}>
                    <Button 
                        variant={continuousShot ? "contained" : "outlined"} 
                        color={continuousShot ? "secondary" : "primary"}
                        fullWidth size="small" 
                        sx={{fontSize: '0.7rem'}} 
                        onClick={toggleContinuousShot}
                    >
                        {continuousShot ? "持续获取中..." : "持续获取截图"}
                    </Button>
                </Grid>
                <Grid item xs={5} sx={{ display: 'flex', alignItems: 'center' }}>
                    <Typography variant="caption" sx={{ mr: 0.5 }}>间隔:</Typography>
                    <CostumedInput defaultValue={intervalMs} onCommit={v => setIntervalMs(Number(v))} width="35px" />
                    <Typography variant="caption" sx={{ ml: 0.5 }}>ms</Typography>
                </Grid>
            </Grid>

            <Divider sx={{ my: 1 }} />

            {/* --- 新增：快捷结束程序卡片区块 --- */}
            <Box sx={{ p: 1, mb: 2, bgcolor: 'rgba(244, 67, 54, 0.05)', border: '1px solid rgba(244, 67, 54, 0.2)', borderRadius: 1 }}>
                <Typography variant="subtitle2" sx={{ color: "#D32F2F", mb: 1, fontWeight: 'bold', fontSize: '0.75rem' }}>结束程序设置</Typography>
                
                <Button 
                    variant="contained" 
                    fullWidth 
                    size="small" 
                    onClick={handleExitClick}
                    disabled={exitBtnState === 2}
                    sx={{ 
                        mb: 1, 
                        fontSize: '0.7rem',
                        fontWeight: 'bold',
                        color: '#fff',
                        bgcolor: exitBtnState === 0 ? '#ED6C02' : (exitBtnState === 1 ? '#EF5350' : '#9E9E9E'), 
                        '&:hover': { bgcolor: exitBtnState === 0 ? '#E65100' : '#D32F2F' } 
                    }}
                >
                    {exitBtnState === 0 && "结束程序"}
                    {exitBtnState === 1 && `再次点击确认结束 (${countdown}s)`}
                    {exitBtnState === 2 && exitMessage}
                </Button>

                <Grid container alignItems="center" justifyContent="space-between" sx={{ mb: 0.5 }}>
                    <Typography variant="body2" sx={{ fontSize: '0.7rem' }}>按键结束程序</Typography>
                    <Switch size="small" checked={isExitEnable} onChange={toggleExitEnable} color="error" />
                </Grid>

                <Grid container alignItems="center" spacing={1}>
                    <Grid item xs={5}>
                        <Typography variant="caption" sx={{ color: '#666' }}>快捷结束绑定键:</Typography>
                    </Grid>
                    <Grid item xs={7}>
                        <Button 
                            variant="outlined" 
                            size="small" 
                            fullWidth
                            disabled={!currentActiveKey}
                            onClick={addExitKey}
                            sx={{ 
                                fontSize: '0.65rem', 
                                p: 0, 
                                borderColor: currentActiveKey ? '#F44336' : '#ccc', 
                                color: currentActiveKey ? '#D32F2F' : '#999' 
                            }}
                        >
                            {currentActiveKey ? `绑定: ${currentActiveKey}` : "按住键盘/手柄按键"}
                        </Button>
                    </Grid>
                    {exitKeys.length > 0 && (
                        <Grid item xs={12}>
                            <Box sx={{ display: 'flex', flexWrap: 'wrap', mt: 0.5 }}>
                                {exitKeys.map((k, idx) => (
                                    <Typography key={idx} variant="caption" sx={{ border: '1px solid #D32F2F', color: '#D32F2F', borderRadius: 1, px: 0.5, mr: 0.5, mb: 0.5, display: 'flex', alignItems: 'center', bgcolor: '#fff' }}>
                                        {k.replace("BTN_", "").replace("KEY_", "")} 
                                        <HighlightOffIcon sx={{ fontSize: '0.8rem', cursor: 'pointer', ml: 0.5, color: '#D32F2F' }} onClick={() => removeExitKey(idx)} />
                                    </Typography>
                                ))}
                            </Box>
                        </Grid>
                    )}
                </Grid>
            </Box>

            <Divider sx={{ my: 1 }} />

            <Badge color="error" variant="dot" invisible={!infoBadgeVisible} sx={{ width: '100%' }}>
                <Button 
                    variant="contained" 
                    fullWidth 
                    size="small" 
                    startIcon={<InfoIcon />} 
                    onClick={handleOpenInfo}
                    sx={{ 
                        bgcolor: '#8bc34a', 
                        color: 'white',
                        '&:hover': { bgcolor: '#7cb342' } 
                    }}
                >
                    查看程序声明
                </Button>
            </Badge>

            <InfoModal open={showInfoModal} onClose={() => setShowInfoModal(false)} />
        </Box>
    );
}