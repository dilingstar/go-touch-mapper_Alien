import React from 'react';
import { 
    Dialog, DialogTitle, DialogContent, IconButton, Typography, Link, Slide, Box 
} from "@mui/material";
import CloseIcon from '@mui/icons-material/Close';
import InfoIcon from '@mui/icons-material/Info';

const Transition = React.forwardRef(function Transition(props, ref) {
    return <Slide direction="up" ref={ref} {...props} />;
});

export default function InfoModal({ open, onClose }) {
    // 通用样式：允许文本选择
    const selectableStyle = { userSelect: 'text', cursor: 'text' };

    return (
        <Dialog
            open={open}
            TransitionComponent={Transition}
            keepMounted
            onClose={onClose}
            aria-describedby="alert-dialog-slide-description"
            PaperProps={{
                sx: {
                    width: '85%',
                    maxWidth: '400px',
                    maxHeight: '80%',
                    borderRadius: 3,
                    background: 'linear-gradient(135deg, #ffffff 0%, #f5f5f5 100%)',
                    boxShadow: '0 8px 32px 0 rgba(31, 38, 135, 0.37)',
                }
            }}
        >
            <DialogTitle sx={{ m: 0, p: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center', bgcolor: '#f0f4c3' }}>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                    <InfoIcon sx={{ color: '#8bc34a', mr: 1 }} />
                    <Typography variant="h6" component="div" sx={{ fontWeight: 'bold', color: '#33691e', fontSize: '1rem' }}>
                        程序说明
                    </Typography>
                </Box>
                <IconButton
                    aria-label="close"
                    onClick={onClose}
                    sx={{
                        color: (theme) => theme.palette.grey[500],
                    }}
                >
                    <CloseIcon />
                </IconButton>
            </DialogTitle>
            <DialogContent dividers sx={{ p: 2 }}>
                <Typography gutterBottom sx={{ fontWeight: 'bold', color: '#d32f2f', ...selectableStyle }}>
                    禁止任何形式的倒卖、收费！
                </Typography>
                <Typography paragraph variant="body2" sx={selectableStyle}>
                    程序完全免费，如果你遇到了收费，那就是本程序被倒卖了。
                    <br />
                    如果你是付费购买的，请立刻联系作者举报，然后找收费者退款。
                </Typography>
                
                <Typography variant="subtitle2" sx={{ mt: 2, fontWeight: 'bold' }}>维护者:</Typography>
                <Typography variant="body2">
                    <Link href="https://b23.tv/rYaWKlV" target="_blank" underline="hover" sx={{ color: '#1976d2', fontWeight: 'bold' }}>dilingstar</Link>
                    {' & '}
                    <Link href="https://b23.tv/QHOrprl" target="_blank" underline="hover" sx={{ color: '#1976d2', fontWeight: 'bold' }}>RiderLty</Link>
                </Typography>

                <Typography variant="subtitle2" sx={{ mt: 2, fontWeight: 'bold' }}>项目地址:</Typography>
                <Typography variant="body2">
                    开源项目: <Link href="https://github.com/RiderLty/go-touch-mapper" target="_blank" underline="hover" sx={{ color: '#1976d2' }}>RiderLty/go-touch-mapper</Link>
                </Typography>
                <Typography variant="body2">
                    Fork项目: <Link href="https://github.com/dilingstar/go-touch-mapper" target="_blank" underline="hover" sx={{ color: '#1976d2' }}>dilingstar/go-touch-mapper</Link>
                </Typography>

                <Typography variant="subtitle2" sx={{ mt: 2, fontWeight: 'bold' }}>联系方式:</Typography>
                <Typography variant="body2" sx={selectableStyle}>
                    Fork QQ群: <Link href="https://qm.qq.com/q/jcTBRjrry0" target="_blank" underline="hover" sx={{ color: '#1976d2' }}>1067729485</Link>
                </Typography>
                <Typography variant="body2" sx={selectableStyle}>
                    作者QQ: 2714637511
                </Typography>

                <Typography variant="subtitle2" sx={{ mt: 2, fontWeight: 'bold' }}>教程:</Typography>
                <Typography variant="body2">
                    <Link href="https://b23.tv/DJHBhgl" target="_blank" underline="hover" sx={{ color: '#1976d2' }}>无需安装apk的手柄键鼠映射工具简单食用指南</Link>
                </Typography>
                <Typography variant="body2" sx={{ mt: 1 }}>
                    <Link href="https://b23.tv/yzq2C9Y" target="_blank" underline="hover" sx={{ color: '#1976d2' }}>安卓手机termux启动gtm程序映射工具小白教程</Link>
                </Typography>
                <Typography variant="body2" sx={{ mt: 1, ...selectableStyle }}>
                    其他教程待更新。如有程序问题bug 请反馈。
                </Typography>

                <Box sx={{ mt: 3, p: 1, bgcolor: '#e3f2fd', borderRadius: 1 }}>
                    <Typography variant="caption" display="block" sx={{ color: '#0d47a1', ...selectableStyle }}>
                        小知识: 在程序运行时，同一网络下，相同地址，其他设备也可进入当前配置页面。
                    </Typography>
                </Box>
            </DialogContent>
        </Dialog>
    );
}