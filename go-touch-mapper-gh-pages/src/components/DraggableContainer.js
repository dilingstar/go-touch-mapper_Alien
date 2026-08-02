import { useEffect, useRef, useState } from "react";
import Paper from '@mui/material/Paper';
import { IconButton } from "@mui/material";
import RemoveIcon from '@mui/icons-material/Remove';
import AddIcon from '@mui/icons-material/Add';
import FullscreenIcon from '@mui/icons-material/Fullscreen';

//可拖动的容器
//标题栏 ：内容
//标题栏可拖动 内容外部传入

export default function DraggableContainer(props) { 

    const mouseDowning = useRef(false)
    const lastPos = useRef([0, 0])

    const [left_top, setLeft_top] = useState([0, 0])
    const left_top_ref = useRef([0, 0])

    // [新增] 维护窗口的缩放比例，默认 1.0
    const [scale, setScale] = useState(1.0)

    const getMovexy = (e) => { 
        if (e.type === "touchmove" || e.type === "touchstart") {
            return [e.touches[0].clientX, e.touches[0].clientY]
        } else if (e.type === "mousedown" || e.type === "mousemove") {
            return [e.clientX, e.clientY]
        } else { 
            return [1,1]
        }
    }

    const onMouseDown = (e) => {
        mouseDowning.current = true
        lastPos.current = getMovexy(e)
    }

    const onMouseMove = (e) => { 
        if (mouseDowning.current) {
            e.preventDefault()
            const offsetX = getMovexy(e)[0] - lastPos.current[0]
            const offsetY = getMovexy(e)[1] - lastPos.current[1]
            lastPos.current = getMovexy(e)
            const new_left_top = [left_top_ref.current[0] + offsetX, left_top_ref.current[1] + offsetY]
            setLeft_top(new_left_top)
            left_top_ref.current = new_left_top
        }
    }
    const onMouseUp = (e) => { 
        mouseDowning.current = false
    }

    useEffect(() => {
        document.onmousemove = onMouseMove
        document.onmouseup = onMouseUp
        
        document.ontouchmove = onMouseMove
        document.ontouchend = onMouseUp
        return () => {
            document.onmousemove = null
            document.onmouseup = null

            document.ontouchmove = null
            document.ontouchend = null
        }
    }, [])

    
    return <Paper
        sx={{
            zIndex: 100, // [V3.2.0] 提升层级，确保覆盖所有可视化元素
            position: 'fixed',
            left: left_top[0],
            top: left_top[1],
            overflow: "hidden",
            borderRadius: "8px",
            boxShadow: "0px 5px 15px rgba(0,0,0,0.3)", // 增加阴影提升悬浮感
            transform: `scale(${scale})`,       // [新增] 应用缩放比例
            transformOrigin: 'top left',        // [新增] 缩放时以左上角为基准点，防止窗口乱飘
        }}
    >
        <div style={{ 
                height: "30px", 
                backgroundColor: "#607D8B", 
                cursor: "move",
                display: "flex",                // [新增] 开启 Flex 布局
                justifyContent: "flex-end",     // [新增] 内部元素靠右对齐
                alignItems: "center",           // [新增] 垂直居中
                paddingRight: "5px"             // [新增] 给右边留一点缝隙
            }}
            onMouseDown={onMouseDown}
            onTouchStart={onMouseDown}
        >
            {/* [新增] 按钮区域。使用 stopPropagation 阻止事件冒泡，这样点击按钮时就不会触发外层的拖拽动作了 */}
            <div 
                onMouseDown={(e) => e.stopPropagation()} 
                onTouchStart={(e) => e.stopPropagation()}
                style={{ display: 'flex' }}
            >
                <IconButton size="small" sx={{ color: 'white', p: 0.2 }} onClick={() => setScale(s => Math.max(0.5, s - 0.25))}>
                    <RemoveIcon fontSize="small" />
                </IconButton>
                <IconButton size="small" sx={{ color: 'white', p: 0.2 }} onClick={() => setScale(s => Math.min(2.0, s + 0.25))}>
                    <AddIcon fontSize="small" />
                </IconButton>
                <IconButton size="small" sx={{ color: 'white', p: 0.2 }} onClick={() => document.body.requestFullscreen()}>
                    <FullscreenIcon fontSize="small" />
                </IconButton>
            </div>
        </div>
        {
            props.children
        }
    </Paper>
}