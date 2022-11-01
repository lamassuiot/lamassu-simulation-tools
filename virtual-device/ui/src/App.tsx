/* eslint-disable */
import React, { useState, useEffect } from "react"
import { Box, Button, ButtonGroup, createTheme, GlobalStyles, Grid, IconButton, keyframes, Paper, Slider, ThemeProvider, Typography } from "@mui/material"
import CachedIcon from "@mui/icons-material/Cached"
import CheckIcon from "@mui/icons-material/Check"
import PriorityHighIcon from "@mui/icons-material/PriorityHigh"
import VpnLockOutlinedIcon from "@mui/icons-material/VpnLockOutlined"
import LockResetOutlinedIcon from "@mui/icons-material/LockResetOutlined"
import MemoryOutlinedIcon from "@mui/icons-material/MemoryOutlined"
import CloseOutlinedIcon from "@mui/icons-material/CloseOutlined"
import MoreHorizIcon from "@mui/icons-material/MoreHoriz"
import ReplayIcon from "@mui/icons-material/Replay"
import DeleteOutlineOutlinedIcon from "@mui/icons-material/DeleteOutlineOutlined"
import moment from "moment"
import { useAppSelector } from "ducks/hooks"
import * as websocketSelector from "ducks/features/websocket/reducer"
import * as deviceManagerSelector from "ducks/features/deviceManager/reducer"
import { useDispatch } from "react-redux"
import { ActionType } from "ducks/features/websocket/actionTypes"
import InspectMode from "components/InspectMode"

const App = () => {
    const dispatch = useDispatch()
    const [showSidebar, setShowSidebar] = useState(false)

    const deviceState = useAppSelector((state: any) => deviceManagerSelector.getDeviceState(state))
    const telemetryData = useAppSelector((state: any) => deviceManagerSelector.getTelemetryData(state))
    const mqttLogs = useAppSelector((state: any) => deviceManagerSelector.getMqttLogs(state))

    const websocketState = useAppSelector((state: any) => websocketSelector.getWebsocketState(state))
    const [selectedIntegration, setSelectedIntegration] = useState("aws")
    const supportedIntegrations = [
        "aws",
        "azure"
    ]

    const [selectedSlotID, setSelectedSlotID] = useState("default")

    const [intervalID, setIntervalID] = useState<any>()
    const [expirationDate, setExpirationDate] = useState("")

    const blinker = keyframes`
    50% {
        opacity: 0;
      }
    `

    const theme = createTheme({
        palette: {
            mode: "dark",
            primary: {
                main: "#25ee32",
                contrastText: "#fff"
            },
            text: {
                primary: "#fff"
            }
        }
    })

    const filteredSlots = deviceState.slots.filter(s => s.id === selectedSlotID)

    let statusColor = "#ED6059"
    if (filteredSlots.length === 1) {
        switch (filteredSlots[0].status) {
            case "PROVISIONED":
                statusColor = "#25ee32"
                break

            case "EXPIRED":
                statusColor = "#ED6059"
                break

            default:
                statusColor = "orange"
                break
        }
    }

    let statusIcon = (
        <CloseOutlinedIcon sx={{ color: "white", fontSize: "70px" }} fontSize="large" />
    )
    if (filteredSlots.length === 1) {
        switch (filteredSlots[0].status) {
            case "PROVISIONED":
                statusIcon = <CheckIcon sx={{ color: "white", fontSize: "120px" }} fontSize="large" />
                break

            case "EXPIRED":
                statusIcon = <CloseOutlinedIcon sx={{ color: "white", fontSize: "120px" }} fontSize="large" />
                break

            default:
                statusIcon = <PriorityHighIcon sx={{ color: "white", fontSize: "120px" }} fontSize="large" />
                break
        }
    }

    const triggerExpriationDateTimer = () => {
        const id = setInterval(() => {
            let newExpirationDate = ""
            if (filteredSlots.length === 1 && moment(filteredSlots[0].expirationDate).valueOf() > 0) {
                newExpirationDate = moment(filteredSlots[0].expirationDate).format("DD/MM/YYYY HH:mm")
                newExpirationDate += " (in "
                const diffMillis = moment(filteredSlots[0].expirationDate).diff(moment())
                const diffDuration = moment.duration(diffMillis)
                if (diffDuration.asSeconds() > 60) {
                    if (diffDuration.asMinutes() > 60) {
                        if (diffDuration.asHours() > 24) {
                            newExpirationDate += Math.round(diffDuration.asDays()) + " days)"
                        } else {
                            newExpirationDate += Math.round(diffDuration.asHours()) + " hours)"
                        }
                    } else {
                        newExpirationDate += Math.round(diffDuration.asMinutes()) + " minutes)"
                    }
                } else {
                    newExpirationDate += Math.round(diffDuration.asSeconds()) + " seconds)"
                }
            }
            if (newExpirationDate !== "") {
                setExpirationDate(newExpirationDate)
            }
        }, 1000)

        setIntervalID(id)
    }

    useEffect(() => {
        if (intervalID) {
            clearInterval(intervalID)
        }
        triggerExpriationDateTimer()
    }, [deviceState])


    return (
        <ThemeProvider theme={theme}>
            <GlobalStyles
                styles={{
                    "*::-webkit-scrollbar": {
                        width: "25px",
                        height: "8px"
                    },
                    "*::-webkit-scrollbar-track": {
                        background: "#555555",
                    },
                    "*::-webkit-scrollbar-thumb": {
                        backgroundColor: "#eeeeee",
                        borderRadius: 50,
                        border: 0,
                        outline: "none"
                    }
                }}
            />
            <Grid container height="100%">
                <Grid item xs={showSidebar ? 9 : 12} height="100%">
                    <Box component={Paper} height="100%" display="flex" flexDirection="column" bgcolor="#1D2025" borderRadius={0}>
                        <Grid component={Paper} spacing={"80px"} sx={{ background: "#1F2933", padding: "10px 28px" }} elevation={4} borderRadius={0} container alignItems="center" justifyContent="space-between">
                            <Grid item xs="auto">
                                <Typography fontSize="35px" fontWeight="400">Device Overview</Typography>
                            </Grid>
                            <Grid item xs="auto" container spacing={2}>
                                <Grid item xs="auto">
                                    <Typography color="#DEE2E7" fontSize="35px" fontWeight="400">{telemetryData.temperature}ºC</Typography>
                                </Grid>
                                <Grid item xs="auto">
                                    <Typography color="#DEE2E7" fontSize="35px" fontWeight="400">/</Typography>
                                </Grid>
                                <Grid item xs="auto">
                                    <Typography color="#DEE2E7" fontSize="35px" fontWeight="400">{telemetryData.humidity}%</Typography>
                                </Grid>
                            </Grid>
                            <Grid item xs container alignItems="center" justifyContent="flex-end" spacing={2}>
                                <Grid item>
                                    <Typography color="#fff" fontSize="25px" fontWeight="400">{websocketState}</Typography>
                                </Grid>
                                {
                                    websocketState === "DISCONNECTED" && (
                                        <Grid item>
                                            <IconButton onClick={() => {
                                                dispatch({
                                                    type: ActionType.WS_RECONNECT
                                                })
                                            }}>
                                                <ReplayIcon />
                                            </IconButton>
                                        </Grid>
                                    )
                                }
                                <Grid item>
                                    <IconButton onClick={() => { setShowSidebar(!showSidebar) }}>
                                        <MoreHorizIcon />
                                    </IconButton>
                                </Grid>
                            </Grid>
                        </Grid>

                        <Box style={{ display: "flex", flexDirection: "column", height: "calc(100% - 120px)", padding: "20px" }}>
                            <Box sx={{ marginBottom: "20px" }}>
                                <Grid container>
                                    <Grid item xs="auto">
                                        <Button variant="contained" sx={{ height: "50px", fontSize: "30px" }} startIcon={<CachedIcon />} onClick={() => {
                                            dispatch({
                                                type: ActionType.WS_SEND_MESSAGE,
                                                value: {
                                                    type: "GEN_NEW_ID",
                                                    message: {
                                                    },
                                                    time: Date.now()
                                                }
                                            })
                                        }}>
                                            Generate Serial Number
                                        </Button>
                                    </Grid>

                                    <Grid item xs container justifyContent="flex-end" spacing={2}>
                                        <Grid item marginRight="30px">
                                            <Button variant="outlined" sx={{ height: "50px", fontSize: "30px" }} startIcon={<MemoryOutlinedIcon />} disabled={deviceState.status !== "WITH_ID"} onClick={() => {
                                                dispatch({
                                                    type: ActionType.WS_SEND_MESSAGE,
                                                    value: {
                                                        type: "GEN_NEW_SLOT",
                                                        message: {
                                                        },
                                                        time: Date.now()
                                                    }
                                                })
                                            }}>
                                                Add New Slot
                                            </Button>
                                        </Grid>
                                        <Grid item>
                                            <Button variant="outlined" sx={{ height: "50px", fontSize: "30px" }} startIcon={<VpnLockOutlinedIcon />} disabled={!(filteredSlots.length === 1 && filteredSlots[0].status === "NEEDS_PROVISIONING")} onClick={() => {
                                                dispatch({
                                                    type: ActionType.WS_SEND_MESSAGE,
                                                    value: {
                                                        type: "ENROLL",
                                                        message: {
                                                            slot_id: selectedSlotID
                                                        },
                                                        time: Date.now()
                                                    }
                                                })
                                            }}>
                                                Issue First Identity
                                            </Button>
                                        </Grid>
                                        <Grid item>
                                            <Button sx={{ height: "50px", fontSize: "30px" }} variant="outlined" disabled={!(filteredSlots.length === 1 && (filteredSlots[0].status === "PROVISIONED" || filteredSlots[0].status === "NEEDS_REENROLLMENT"))} startIcon={<LockResetOutlinedIcon />} onClick={() => {
                                                dispatch({
                                                    type: ActionType.WS_SEND_MESSAGE,
                                                    value: {
                                                        type: "REENROLL",
                                                        message: {
                                                            slot_id: selectedSlotID
                                                        },
                                                        time: Date.now()
                                                    }
                                                })
                                            }}>
                                                Renew Identity
                                            </Button>
                                        </Grid>
                                    </Grid>
                                </Grid>
                            </Box>

                            <Box sx={{ marginBottom: "20px" }}>
                                <Grid container spacing={2}>
                                    <Grid item xs={6} container spacing={2}>
                                        <Grid item xs={12}>
                                            <Box bgcolor="#1F2933" component={Paper} padding="25px">
                                                <Typography color="#B2B3B7" fontSize="25px" fontWeight="400">Device Serial Number</Typography>
                                                <Typography color="#DEE2E7" fontSize="28px" fontWeight="400" display="inline-block">{deviceState.serialNumber}</Typography>
                                            </Box>
                                        </Grid>
                                        <Grid item xs={12}>
                                            <Box bgcolor="#1F2933" component={Paper} padding="25px" display="flex" alignItems="center" justifyContent="center" flexDirection="column" height="calc(100% - 50px)">
                                                <Typography color="#B2B3B7" fontSize="25px" fontWeight="400">Status</Typography>
                                                <Grid container bgcolor={statusColor} alignItems="center" justifyContent="center" width="200px" height="200px" borderRadius="200px" margin="25px 0">
                                                    {
                                                        filteredSlots.length === 1 && (filteredSlots[0].status === "PENDING_PROVISIONING" || filteredSlots[0].status === "PENDING_REENROLLMENT")
                                                            ? (
                                                                <>
                                                                    <Grid item>
                                                                        <Box width="40px" height="40px" borderRadius="40px" bgcolor="white" margin="0 4px" sx={{
                                                                            animation: `${blinker} 1.5s linear infinite`
                                                                        }} />
                                                                    </Grid>
                                                                    <Grid item>
                                                                        <Box width="40px" height="40px" borderRadius="40px" bgcolor="white" margin="0 4px" sx={{
                                                                            animation: `${blinker} 1.5s linear infinite`,
                                                                            animationDelay: "0.3s"
                                                                        }} />
                                                                    </Grid>
                                                                    <Grid item>
                                                                        <Box width="40px" height="40px" borderRadius="40px" bgcolor="white" margin="0 4px" sx={{
                                                                            animation: `${blinker} 1.5s linear infinite`,
                                                                            animationDelay: "0.6s"
                                                                        }} />
                                                                    </Grid>
                                                                </>
                                                            )
                                                            : (
                                                                <>
                                                                    {statusIcon}
                                                                </>
                                                            )
                                                    }
                                                </Grid>
                                                {
                                                    filteredSlots.length === 1
                                                        ? (
                                                            <Typography color="#DEE2E7" fontSize="30px" fontWeight="400">{filteredSlots[0].status}</Typography>
                                                        )
                                                        : (
                                                            <Typography color="#DEE2E7" fontSize="30px" fontWeight="400">-</Typography>
                                                        )
                                                }
                                            </Box>
                                        </Grid>

                                    </Grid>

                                    <Grid item xs container>
                                        <Box bgcolor="#1F2933" component={Paper} padding="25px" flex="1">
                                            <Grid container spacing={2}>
                                                <Grid item xs={12}>
                                                    {
                                                        filteredSlots.length === 1
                                                            ? (
                                                                <ButtonGroup aria-label="medium button group">
                                                                    {
                                                                        deviceState.slots.map((slot, idx) => (
                                                                            <Button key={idx} sx={{ height: "50px", fontSize: "30px" }} variant={selectedSlotID === slot.id ? "contained" : "outlined"} onClick={() => {
                                                                                setSelectedSlotID(slot.id)
                                                                            }}>Slot {slot.id}</Button>
                                                                        ))
                                                                    }
                                                                </ButtonGroup>
                                                            )
                                                            : (
                                                                <Typography color="#B2B3B7" fontSize="25px" fontWeight="400">{"Can't display slot info"}</Typography>
                                                            )
                                                    }
                                                </Grid>
                                                {
                                                    filteredSlots.length === 1 &&
                                                    (
                                                        <>
                                                            <Grid item xs={12}>
                                                                <Typography color="#B2B3B7" fontSize="25px" fontWeight="400">Certificate Serial Number</Typography>
                                                                <Typography color="#DEE2E7" fontSize="28px" fontWeight="400">{filteredSlots[0].serialNumber}</Typography>
                                                            </Grid>
                                                            <Grid item xs={12}>
                                                                <Typography color="#B2B3B7" fontSize="25px" fontWeight="400">Issuer Certificate Authority</Typography>
                                                                <Typography color="#DEE2E7" fontSize="28px" fontWeight="400">{filteredSlots[0].issuingCA}</Typography>
                                                            </Grid>
                                                            <Grid item xs={12}>
                                                                <Typography color="#B2B3B7" fontSize="25px" fontWeight="400">Expiration Date</Typography>
                                                                <Typography color="#DEE2E7" fontSize="28px" fontWeight="400">{expirationDate}</Typography>
                                                            </Grid>
                                                        </>
                                                    )

                                                }

                                            </Grid>
                                        </Box>
                                    </Grid>
                                </Grid>
                            </Box>

                            <Box bgcolor="#1F2933" component={Paper} sx={{ height: "100%", flexGrow: 1, overflowY: "auto", padding: "20px", display: "flex", flexDirection: "column" }}>
                                <Box sx={{ width: "100%", marginBottom: "20px" }}>
                                    <Grid container>
                                        <Grid item xs>
                                            <ButtonGroup size="medium" aria-label="medium button group">
                                                {
                                                    supportedIntegrations.map((integration, idx) => (
                                                        <Button key={idx} sx={{ height: "50px", fontSize: "30px" }} variant={selectedIntegration === integration ? "contained" : "outlined"} onClick={() => {
                                                            setSelectedIntegration(integration)
                                                        }}>{integration}</Button>
                                                    ))
                                                }
                                            </ButtonGroup>
                                        </Grid>
                                        <Grid item xs="auto" container alignItems="center" spacing={1}>
                                            <Grid item>
                                                <Typography>{deviceState.mqttConnected ? "CONNECTED" : "DISCONNECTED"}</Typography>
                                            </Grid>
                                            <Grid item>
                                                <Button sx={{ height: "50px", fontSize: "30px" }} variant="outlined" startIcon={<ReplayIcon />} onClick={() => {
                                                    if (deviceState.mqttConnected) {
                                                        dispatch({
                                                            type: ActionType.WS_SEND_MESSAGE,
                                                            value: {
                                                                type: "MQTT_DISCONNECT",
                                                                message: {},
                                                                time: Date.now()
                                                            }
                                                        })
                                                    } else {
                                                        dispatch({
                                                            type: ActionType.WS_SEND_MESSAGE,
                                                            value: {
                                                                type: "MQTT_CONNECT",
                                                                message: {
                                                                    slot_id: selectedSlotID,
                                                                    provider: selectedIntegration
                                                                },
                                                                time: Date.now()
                                                            }
                                                        })
                                                    }
                                                }}
                                                >
                                                    {deviceState.mqttConnected ? "Disconnect" : "Connect"}
                                                </Button>
                                            </Grid>
                                        </Grid>
                                    </Grid>
                                </Box>


                                <Box bgcolor="#151515" borderRadius="5px" sx={{ height: "100%", flexGrow: 1, overflowY: "auto", padding: "10px" }}>
                                    <Grid container spacing={2} sx={{}}>
                                        {
                                            mqttLogs.map((log, idx) => {
                                                let color = "#25ee32"
                                                if (log.type === "ERROR") {
                                                    color = "#ee3125"
                                                } else if (log.type === "INFO") {
                                                    color = "#1DA8C3"
                                                }

                                                return (
                                                    <Grid item xs={12} container key={idx}>
                                                        <Grid item xs={12}>
                                                            <Typography color={color} fontSize="18px" fontWeight="400">{log.title}</Typography>
                                                        </Grid>
                                                        <Grid item xs={12}>
                                                            <Typography color="#B2B3B7" fontSize="18px" fontWeight="400" fontFamily="monospace">{log.message}</Typography>
                                                        </Grid>
                                                    </Grid>
                                                )
                                            })
                                        }
                                    </Grid>
                                </Box>
                            </Box>
                        </Box>
                    </Box >
                </Grid >
                {
                    showSidebar && (
                        <InspectMode />
                    )
                }
            </Grid >

        </ThemeProvider >
    )
}

export default App
