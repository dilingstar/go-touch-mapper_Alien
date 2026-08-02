import { useEffect, useState } from "react";
import { Input } from "@mui/material";

// Version: V3.4.5

const UploadButton = ({ onClick }) => {
    return <button
        style={{
            position: 'absolute',
            width: '200px',
            height: '80px',
            left: '50%',
            marginLeft: '-105px',
            top: 'calc(50% - 100px)',
            borderRadius: '50px',
            border: "5px solid #00b894",
            transition: ".25s",
            fontSize: '24px',
            background: "#2C3A47",
            color: "white",
        }}
        onClick={onClick}>上传图片</button>
}

const UploadButtonJIETU = ({ onClick }) => {
    return <button
        style={{
            position: 'absolute',
            width: '200px',
            height: '80px',
            left: '50%',
            marginLeft: '-105px',
            top: '50%',
            borderRadius: '50px',
            border: "5px solid #00b894",
            transition: ".25s",
            fontSize: '24px',
            background: "#2C3A47",
            color: "white",
        }}
        onClick={onClick}>屏幕截图</button>
}

const UploadButton5s = ({ onClick }) => {
    return <button
        style={{
            position: 'absolute',
            width: '200px',
            height: '80px',
            left: '50%',
            marginLeft: '-105px',
            top: 'calc(50% + 100px)',
            borderRadius: '50px',
            border: "5px solid #00b894",
            transition: ".25s",
            fontSize: '24px',
            background: "#2C3A47",
            color: "white",
        }}
        onClick={onClick}>5s后截图</button>
}

const ConnectorLine = ({ x1, y1, x2, y2, color, type }) => {
    const length = Math.sqrt((x2 - x1) ** 2 + (y2 - y1) ** 2);
    const angle = Math.atan2(y2 - y1, x2 - x1) * 180 / Math.PI;
    
    if (length < 5) return null;

    const commonStyle = {
        position: 'absolute',
        top: y1,
        left: x1,
        height: 1,
        transformOrigin: "0 0",
        pointerEvents: "none",
        zIndex: 1
    };

    if (type === "dashed") {
        return <div style={{ ...commonStyle, width: length, transform: `rotate(${angle}deg)`, borderTop: `1px dashed ${color}`, backgroundColor: "transparent" }} />;
    } else if (type === "gap") {
        const segLen = Math.min(length * 0.2, 20); 
        return (
            <>
                <div style={{ ...commonStyle, width: segLen, transform: `rotate(${angle}deg)`, backgroundColor: color }} />
                <div style={{ 
                    position: 'absolute', top: y2, left: x2, height: 1, width: segLen,
                    transformOrigin: "0 0", pointerEvents: "none", zIndex: 1,
                    transform: `rotate(${angle + 180}deg)`, backgroundColor: color 
                }} />
            </>
        );
    }
    return <div style={{ ...commonStyle, width: length, transform: `rotate(${angle}deg)`, backgroundColor: color }} />;
}

const EyeIcon = ({ size, color }) => (
    <div style={{
        width: size * 0.6,
        height: size * 0.4,
        border: `1px solid ${color}`,
        transform: "rotate(45deg) skew(15deg, 15deg)",
        position: "absolute",
        top: "50%",
        left: "50%",
        marginTop: -size * 0.2,
        marginLeft: -size * 0.3,
    }} />
);

const FixedIcon = ({ x, y, size, type, text, random_radius }) => {
    let baseSize = size || 28;
    if (random_radius && random_radius > 0) {
        baseSize = random_radius * 2;
    } else if (random_radius !== undefined) {
        baseSize = 10; 
    }

    let style = {
        width: baseSize,
        height: baseSize,
        borderRadius: baseSize,
        backgroundColor: "#d90051",
        border: "none",
        color: "white"
    };
    
    let centerEl = null;
    const isCombo = text && text.includes("+") && !type.startsWith("MACRO");

    switch (type) {
        case "PRESS":
        case "CLICK":
        case "AUTO_FIRE":
        case "SEQUENTIAL_PRESS":
            style.backgroundColor = "rgba(0,0,0,0.1)";
            style.border = "1px solid #FF5722";
            break;
        case "DRAG":
            style.backgroundColor = "transparent";
            style.border = "1px solid #F48FB1";
            centerEl = <div style={{width: 4, height: 4, borderRadius: 2, backgroundColor: "black"}} />;
            break;
        case "MULT_PRESS":
            style.backgroundColor = "rgba(255,235,59,0.1)";
            style.border = "1px solid #FF5722";
            break;
        case "SYNC_VIEW_RESET":
        case "CLICK_VIEW_RESET":
            style.backgroundColor = "rgba(205,220,57,0.2)";
            style.border = "1px solid #F44336";
            centerEl = <EyeIcon size={baseSize} color="white" />;
            break;
        case "BACKPACK_TOGGLE":
        case "SYNC_BACKPACK":
        case "SMART_TOGGLE": // [V3.4.5] 兼容智能切出紫色UI
            style.backgroundColor = "rgba(156,39,176,0.2)"; // 更明显的紫色半透明
            style.border = "1px solid #9C27B0";
            break;
        case "PRESS_RELEASE_CLICK":
            style.backgroundColor = "rgba(0,0,0,0.1)";
            style.border = "1px solid #FF5722";
            break;
        case "MACRO_CLICK":
        case "MACRO_SWIPE":
            style.backgroundColor = "rgba(200, 200, 200, 0.2)";
            style.border = "1px solid #B0BEC5";
            style.color = "#CFD8DC";
            if (type === "MACRO_SWIPE") {
                style.border = "1px dashed #B0BEC5";
                centerEl = <div style={{width: 4, height: 4, borderRadius: 2, backgroundColor: "#90A4AE"}} />;
            }
            break;
        default:
            if (isCombo) {
                style.width = baseSize * 2.0; 
                style.borderRadius = baseSize / 2;
                style.backgroundColor = "rgba(0,0,0,0.1)";
                style.border = "1px solid #FF5722";
            }
            break;
    }

    if (isCombo) {
        style.width = baseSize * 2.0; 
        style.borderRadius = baseSize / 2;
    }

    return <div
        style={{
            position: 'absolute',
            left: x,
            top: y,
            marginLeft: style.width / -2,
            marginTop: style.height / -2,
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
            pointerEvents: "none",
            zIndex: 2, 
            ...style
        }}
    >
        {centerEl}
        <span style={{
            position: "absolute",
            color: style.color || "white",
            textShadow: "1px 1px 2px black",
            fontSize: "12px",
            whiteSpace: "pre-line", 
            textAlign: "center",
            zIndex: 3
        }}>
            {text}
        </span>
    </div>
}

const GroupFixedIcon = ({ pos_s, type, text, size, random_radius }) => {
    let lineColor = "#FF5722";
    let lineType = "solid";

    switch (type) {
        case "DRAG":
            lineColor = "#F48FB1";
            lineType = "solid";
            break;
        case "MULT_PRESS":
        case "SEQUENTIAL_PRESS":
            lineColor = "#FF5722";
            lineType = "dashed";
            break;
        case "BACKPACK_TOGGLE":
        case "SYNC_BACKPACK":
        case "SMART_TOGGLE": // [V3.4.5] 紫色断点连线
            lineColor = "#9C27B0";
            lineType = "gap";
            break;
        case "PRESS_RELEASE_CLICK":
            lineColor = "#FF5722";
            lineType = "gap";
            break;
        case "MACRO_SWIPE":
            lineColor = "#B0BEC5"; 
            lineType = "solid";
            break;
    }

    return <div>
        {pos_s.map((pos, index) => {
            if (index === pos_s.length - 1) return null;
            const nextPos = pos_s[index + 1];
            return <ConnectorLine 
                key={`l_${index}`} 
                x1={pos[0]} y1={pos[1]} 
                x2={nextPos[0]} y2={nextPos[1]} 
                color={lineColor} 
                type={lineType} 
            />
        })}

        {pos_s.map((pos, index) => (
            <FixedIcon
                key={index}
                x={pos[0]}
                y={pos[1]}
                size={size}
                type={type}
                text={index === 0 ? text : (type === "MACRO_SWIPE" ? "" : `${text}_${index}`)}
                random_radius={random_radius}
            />
        ))}
    </div>
}

const CostumedInput = ({ defaultValue, width, onCommit, all, disabled, type, inputProps }) => {
    const [value, setValue] = useState(defaultValue)
    useEffect(() => {
        setValue(defaultValue)
    }, [defaultValue])

    return <Input
        sx={{ width: width || "40px" }}
        disabled={disabled}
        type={type}
        inputProps={inputProps}
        value={value}
        onChange={(e) => {
            setValue(e.target.value)
        }}
        onFocus={(e) => {
            window.stopPreventDefault = true
        }}
        onBlur={(e) => {
            window.stopPreventDefault = false
            onCommit && onCommit(all ? value : Number(value))
        }}
        onKeyDown={(e) => {
            if (e.key === "Enter") {
                window.stopPreventDefault = false
                e.target.blur() 
            }
        }}
    />
}

const CostumedStringInput = ({ defaultValue, width, onCommit }) => {
    const [value, setValue] = useState(defaultValue || "")
    useEffect(() => {
        setValue(defaultValue || "");
    }, [defaultValue]);

    return <Input
        sx={{ width: width || "80px" }}
        value={value}
        onChange={(e) => {
            setValue(e.target.value)
        }}
        onFocus={(e) => {
            window.stopPreventDefault = true
        }}
        onBlur={(e) => {
            window.stopPreventDefault = false
            onCommit && onCommit(value) 
        }}
        onKeyDown={(e) => {
            if (e.key === "Enter") {
                e.target.blur(); 
            }
        }}
    />
}

const WheelShow = ({ x, y, range, shift_range, bezier_enable, random_point_enable, random_start_radius, random_end_radius, random_shift_end_radius }) => {
    const radius = range * 2
    const shift_radius = shift_range * 2

    const getControlPoints = (r) => {
        const points = [];
        for (let i = 0; i < 8; i++) {
            const angle = i * 45 * (Math.PI / 180);
            points.push({
                left: x + r * Math.cos(angle),
                top: y + r * Math.sin(angle)
            });
        }
        return points;
    }

    const mainPoints = bezier_enable ? getControlPoints(range) : [];
    const shiftPoints = (bezier_enable && shift_range !== 0) ? getControlPoints(shift_range) : [];

    return <div>
        {random_point_enable && <div style={{
            position: 'absolute', left: x, top: y,
            width: random_start_radius * 2, height: random_start_radius * 2,
            borderRadius: "50%", marginLeft: -random_start_radius, marginTop: -random_start_radius,
            backgroundColor: "rgba(255, 152, 0, 0.3)", border: "1px solid #E65100", pointerEvents: "none",
        }} />}

        <div style={{
            position: 'absolute', left: x, top: y, width: 16, height: 16, borderRadius: 16,
            marginLeft: -8, marginTop: -8, backgroundColor: "#2196F3", pointerEvents: "none",
        }} />

        <div style={{
            position: 'absolute', left: x, top: y, width: radius, height: radius, borderRadius: radius,
            marginLeft: radius / -2 - 4, marginTop: radius / -2 - 4, border: "4px solid #2196F3", pointerEvents: "none",
        }} />

        {shift_range !== 0 && <div style={{
            position: 'absolute', left: x, top: y, width: shift_radius, height: shift_radius, borderRadius: shift_radius,
            marginLeft: shift_radius / -2 - 4, marginTop: shift_radius / -2 - 4, border: "4px solid #512DA8", pointerEvents: "none",
        }} />}

        {bezier_enable && mainPoints.map((pt, idx) => (
            <div key={`mp_${idx}`}>
                <div style={{
                    position: 'absolute', left: pt.left, top: pt.top, width: 8, height: 8, borderRadius: 4,
                    marginLeft: -4, marginTop: -4, backgroundColor: "#0D47A1", pointerEvents: "none",
                }} />
                {random_point_enable && <div style={{
                    position: 'absolute', left: pt.left, top: pt.top,
                    width: random_end_radius * 2, height: random_end_radius * 2,
                    borderRadius: "50%", marginLeft: -random_end_radius, marginTop: -random_end_radius,
                    backgroundColor: "rgba(255, 152, 0, 0.3)", border: "1px solid #E65100", pointerEvents: "none",
                }} />}
            </div>
        ))}

        {bezier_enable && shiftPoints.map((pt, idx) => (
            <div key={`sp_${idx}`}>
                <div style={{
                    position: 'absolute', left: pt.left, top: pt.top, width: 8, height: 8, borderRadius: 4,
                    marginLeft: -4, marginTop: -4, backgroundColor: "#311B92", pointerEvents: "none",
                }} />
                {random_point_enable && <div style={{
                    position: 'absolute', left: pt.left, top: pt.top,
                    width: random_shift_end_radius * 2, height: random_shift_end_radius * 2,
                    borderRadius: "50%", marginLeft: -random_shift_end_radius, marginTop: -random_shift_end_radius,
                    backgroundColor: "rgba(156, 39, 176, 0.3)", border: "1px solid #7B1FA2", pointerEvents: "none",
                }} />}
            </div>
        ))}
    </div>
}

const ViewShow = ({ x, y }) => {
    return <div>
        <div style={{
            position: 'absolute', left: 0, top: y, width: "100vw", height: 1,
            backgroundColor: "#d90051", pointerEvents: "none",
        }} />
        <div style={{
            position: 'absolute', left: x, top: 0, height: "100vh", width: 1,
            backgroundColor: "#d90051", pointerEvents: "none",
        }} />
        <div style={{
            position: 'absolute', left: x, top: y, width: 4, height: 4, borderRadius: 2,
            marginLeft: -2, marginTop: -2, backgroundColor: "#d90051", pointerEvents: "none",
        }} />
    </div>
}

const ViewResetRadiusShow = ({ x, y, radius, enable, random_radius, random_enable }) => {
    if (!enable && !random_enable) return null; // [修改1] 只要任意一个开着，就不拦截渲染
    const diameter = radius * 2;
    return <div>
        {random_enable && <div style={{
            position: 'absolute', left: x, top: y,
            width: random_radius * 2, height: random_radius * 2,
            borderRadius: "50%", marginLeft: -random_radius, marginTop: -random_radius,
            backgroundColor: "rgba(255, 87, 34, 0.1)", border: "1px solid #FF5722", pointerEvents: "none",
        }} />}
        
        {enable && <div style={{ // [修改2] 用 {enable && ...} 将主半径圈起来，让它自己独立判断显示
            position: 'absolute', left: x, top: y, width: diameter, height: diameter,
            borderRadius: "50%", marginLeft: diameter / -2 - 1, marginTop: diameter / -2 - 1,
            border: "2px dashed #FFEB3B", pointerEvents: "none", opacity: 0.7,
        }} />}
    </div>
}

const ScrollSliderShow = ({ x, y, lengthUp, lengthDown, enable, random_start_radius, random_enable }) => {
    if (!enable) return null;
    const totalHeight = lengthUp + lengthDown;
    return <div>
        {random_enable && <div style={{
            position: 'absolute', left: x, top: y,
            width: random_start_radius * 2, height: random_start_radius * 2,
            borderRadius: "50%", marginLeft: -random_start_radius, marginTop: -random_start_radius,
            backgroundColor: "rgba(255, 152, 0, 0.3)", border: "1px solid #E65100", pointerEvents: "none", zIndex: 0,
        }} />}

        <div style={{
            position: 'absolute', left: x, top: y, width: 12, height: 12, borderRadius: 6,
            marginLeft: -6, marginTop: -6, backgroundColor: "#FF9800", pointerEvents: "none", zIndex: 1,
        }} />
        <div style={{
            position: 'absolute', left: x, top: y - lengthUp, width: 10, height: totalHeight,
            marginLeft: -5, backgroundColor: "#03A9F4", opacity: 0.5, pointerEvents: "none", borderRadius: 5,
        }} />
    </div>
}

const VMouseResetPosShow = ({ x, y, enable }) => {
    if (!enable) return null;
    const color = "#00E676";
    return <div>
        <div style={{ position: 'absolute', left: 0, top: y, width: "100vw", height: 1, backgroundColor: color, pointerEvents: "none", opacity: 0.8 }} />
        <div style={{ position: 'absolute', left: x, top: 0, height: "100vh", width: 1, backgroundColor: color, pointerEvents: "none", opacity: 0.8 }} />
        <div style={{ position: 'absolute', left: x, top: y, width: 32, height: 32, borderRadius: 16, marginLeft: -16, marginTop: -16, border: `2px solid ${color}`, backgroundColor: "transparent", pointerEvents: "none" }} />
        <div style={{ position: 'absolute', left: x, top: y, width: 4, height: 4, borderRadius: 2, marginLeft: -2, marginTop: -2, backgroundColor: color, pointerEvents: "none" }} />
    </div>
}

export {
    UploadButton,
    UploadButtonJIETU,
    UploadButton5s,
    FixedIcon,
    GroupFixedIcon,
    CostumedInput,
    CostumedStringInput,
    WheelShow,
    ViewShow,
    ViewResetRadiusShow,
    ScrollSliderShow,
    VMouseResetPosShow,
}