/* eslint-disable */
import React, { useState, useEffect } from "react"
import { Box, Button, createTheme, Grid, MenuItem, IconButton, Paper, Select, styled, Switch, ThemeProvider, Typography, TextField, Divider } from "@mui/material"
import PriorityHighIcon from "@mui/icons-material/PriorityHigh"
import DeleteOutlineOutlinedIcon from "@mui/icons-material/DeleteOutlineOutlined"
import MoreHorizIcon from "@mui/icons-material/MoreHoriz"
import ReplayIcon from "@mui/icons-material/Replay"
import { useAppSelector } from "ducks/hooks"
import moment from "moment"
import * as websocketSelector from "ducks/features/websocket/reducer"
import * as enrollProcesorSelector from "ducks/features/enrollProcesor/reducer"
import * as dmsSelector from "ducks/features/dms/reducer"
import { useDispatch } from "react-redux"
import { ActionType } from "ducks/features/websocket/actionTypes"
import { ArrowSeparator } from "components/ArrowSeparator"
import CheckIcon from "@mui/icons-material/Check"

const Android12Switch = styled(Switch)(({ theme }) => ({
    padding: 8,
    "& .MuiSwitch-track": {
        borderRadius: 22 / 2,
        "&:before, &:after": {
            content: "\"\"",
            position: "absolute",
            top: "50%",
            transform: "translateY(-50%)",
            width: 16,
            height: 16
        },
        "&:before": {
            backgroundImage: `url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" height="16" width="16" viewBox="0 0 24 24"><path fill="${encodeURIComponent(
                theme.palette.getContrastText(theme.palette.primary.main)
            )}" d="M21,7L9,19L3.5,13.5L4.91,12.09L9,16.17L19.59,5.59L21,7Z"/></svg>')`,
            left: 12
        },
        "&:after": {
            backgroundImage: `url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" height="16" width="16" viewBox="0 0 24 24"><path fill="${encodeURIComponent(
                theme.palette.getContrastText(theme.palette.primary.main)
            )}" d="M19,13H5V11H19V13Z" /></svg>')`,
            right: 12
        }
    },
    "& .MuiSwitch-thumb": {
        boxShadow: "none",
        width: 16,
        height: 16,
        margin: 2
    }
}))

const App = () => {
    const dispatch = useDispatch()
    const websocketState = useAppSelector((state: any) => websocketSelector.getWebsocketState(state))
    const websocketMessages = useAppSelector((state: any) => websocketSelector.getMessages(state))
    const enrollmentProcesState = useAppSelector((state: any) => enrollProcesorSelector.getState(state))
    const dmsState = useAppSelector((state: any) => dmsSelector.getState(state))

    const [registerDMSName, setRegisterDMSName] = useState("")
    const [operatorUsername, setOperatorUsername] = useState("")
    const [operatorPassword, setOperatorPassword] = useState("")

    const [selectedCAForEnrollment, setSelectedCAForEnrollment] = useState<string | undefined>()

    const [showSidebar, setShowSidebar] = useState(false)

    console.log(enrollmentProcesState)
    console.log(dmsState)

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

    useEffect(() => {
        dispatch({
            type: ActionType.WS_SEND_MESSAGE,
            value: {
                type: "GET_CFG",
                time: Date.now()
            }
        })
    }, [])

    useEffect(() => {
        if (selectedCAForEnrollment !== undefined) {
            dispatch({
                type: ActionType.WS_SEND_MESSAGE,
                value: {
                    type: "CFG_SELECTED_CA_FOR_ENROLLMENT",
                    time: Date.now(),
                    message: {
                        selected_ca: selectedCAForEnrollment
                    }
                }
            })
        }
    }, [selectedCAForEnrollment])

    useEffect(() => {
        if (dmsState.selectedCA !== undefined) {
            setSelectedCAForEnrollment(dmsState.selectedCA)
        }
    }, [dmsState])

    let statusIcon = (
        <PriorityHighIcon sx={{ color: "white" }} fontSize="large" />
    )
    let statusColor
    switch (dmsState.status) {
        case "IDLE":
            statusIcon = <CheckIcon sx={{ color: "white" }} fontSize="large" />
            statusColor = theme.palette.primary.main
            break

        default:
            statusIcon = <PriorityHighIcon sx={{ color: "white" }} fontSize="large" />
            statusColor = "orange"
            break
    }

    return (
        <ThemeProvider theme={theme}>
            <Grid container height="100%">
                <Grid item xs={showSidebar ? 9 : 12} height="100%">
                    <Box component={Paper} height="100%" display="flex" flexDirection="column" bgcolor="#1D2025" borderRadius={0}>
                        <Grid component={Paper} sx={{ background: "#1F2933", padding: "10px 15px" }} elevation={4} borderRadius={0} container spacing={4} alignItems="center" justifyContent="space-between">
                            <Grid item xs="auto">
                                <Typography fontSize="25px" fontWeight="400">Device Manufacturing System Overview</Typography>
                            </Grid>
                            <Grid item xs="auto">
                                <Typography fontSize="23px" fontWeight="400" color="#aaa">{dmsState.name}</Typography>
                            </Grid>
                            <Grid item xs="auto">
                                <Typography fontSize="23px" fontWeight="400" color="#aaa">https://dev-lamassu.zpd.ikerlan.es</Typography>
                            </Grid>
                            <Grid item xs container alignItems="center" justifyContent="flex-end" spacing={2}>
                                <Grid item>
                                    <Typography color="#fff" fontSize="23px" fontWeight="400">{websocketState}</Typography>
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

                        <Grid
                            container
                            spacing="20px"
                            padding="40px"
                            flex="1"
                            sx={{ overflowY: "auto" }}
                        >
                            {
                                dmsState.status === "EMPTY"
                                    ? (
                                        <Grid item xs={3} container>
                                            <Box bgcolor="#1F2933" component={Paper} padding="30px 40px 40px 40px" flex="1">
                                                <Grid container spacing="80px">
                                                    <Grid item xs={12} container spacing="30px">
                                                        <Grid item container flexDirection="column">
                                                            <Grid item>
                                                                <Typography color="#B2B3B7" fontSize="23px" fontWeight="400">Operator Username</Typography>
                                                            </Grid>
                                                            <Grid item>
                                                                <TextField label="" variant="standard" value={operatorUsername} fullWidth onChange={(ev) => { setOperatorUsername(ev.target.value) }} />
                                                            </Grid>
                                                        </Grid>
                                                        <Grid item container flexDirection="column">
                                                            <Grid item>
                                                                <Typography color="#B2B3B7" fontSize="23px" fontWeight="400">Operator Password</Typography>
                                                            </Grid>
                                                            <Grid item>
                                                                <TextField label="" variant="standard" value={operatorPassword} fullWidth onChange={(ev) => { setOperatorPassword(ev.target.value) }} />
                                                            </Grid>
                                                        </Grid>
                                                    </Grid>
                                                    <Grid item xs={12} container spacing="30px">
                                                        <Grid item container flexDirection="column">
                                                            <Grid item>
                                                                <Typography color="#B2B3B7" fontSize="23px" fontWeight="400">DMS Name</Typography>
                                                            </Grid>
                                                            <Grid item>
                                                                <TextField label="" variant="standard" fullWidth value={registerDMSName} onChange={(ev) => setRegisterDMSName(ev.target.value)} />
                                                            </Grid>
                                                        </Grid>
                                                        <Grid item container flexDirection="column">
                                                            <Grid item>
                                                                <Button variant="contained" onClick={() => {
                                                                    dispatch({
                                                                        type: ActionType.WS_SEND_MESSAGE,
                                                                        value: {
                                                                            type: "CFG",
                                                                            time: Date.now(),
                                                                            message: {
                                                                                operator_username: operatorUsername,
                                                                                operator_password: operatorPassword,
                                                                                dms_name: registerDMSName
                                                                            }
                                                                        }
                                                                    })
                                                                }}>Register</Button>
                                                            </Grid>
                                                        </Grid>
                                                    </Grid>
                                                </Grid>
                                            </Box>
                                        </Grid>

                                    )
                                    : (
                                        <Grid item container direction={"column"} spacing={"40px"} sx={{ padding: "60px 0px" }}>
                                            <Grid item container spacing={"40px"}>
                                                <Grid item xs={6}>
                                                    <Box bgcolor="#1F2933" component={Paper} padding="20px" flex="1">
                                                        <Grid container spacing={2}>
                                                            <Grid item xs={12} container>
                                                                <Grid item xs>
                                                                    <Typography color="#B2B3B7" fontSize="23px" fontWeight="400">Enroll Devices with CA</Typography>
                                                                </Grid>
                                                                <Grid item xs="auto">
                                                                    <Select
                                                                        value={selectedCAForEnrollment}
                                                                        onChange={(ev) => { setSelectedCAForEnrollment(ev.target.value) }}
                                                                        size="medium"
                                                                        variant="standard"
                                                                    >
                                                                        {
                                                                            dmsState.authorizedCAs.map((ca, index) => (
                                                                                <MenuItem key={index} value={ca}>{ca}</MenuItem>
                                                                            ))
                                                                        }
                                                                    </Select>
                                                                </Grid>
                                                            </Grid>
                                                            <Grid item xs={12} container>
                                                                <Grid item xs>
                                                                    <Typography color="#B2B3B7" fontSize="23px" fontWeight="400">Requires manual approval while enrolling</Typography>
                                                                </Grid>
                                                                <Grid item xs="auto">
                                                                    <Android12Switch onChange={(ev, checked) => {
                                                                        dispatch({
                                                                            type: ActionType.WS_SEND_MESSAGE,
                                                                            value: {
                                                                                type: "CFG_AUTO_ENROLLMENT",
                                                                                message: {
                                                                                    auto_enroll: checked
                                                                                },
                                                                                time: Date.now()
                                                                            }
                                                                        })
                                                                    }} />
                                                                </Grid>
                                                            </Grid>
                                                            <Grid item xs={12} container>
                                                                <Grid item xs>
                                                                    <Typography color="#B2B3B7" fontSize="23px" fontWeight="400">Requires manual approval while transfering certificate</Typography>
                                                                </Grid>
                                                                <Grid item xs="auto">
                                                                    <Android12Switch onChange={(ev, checked) => {
                                                                        dispatch({
                                                                            type: ActionType.WS_SEND_MESSAGE,
                                                                            value: {
                                                                                type: "CFG_AUTO_TRANSFER",
                                                                                message: {
                                                                                    auto_transfer: checked
                                                                                },
                                                                                time: Date.now()
                                                                            }
                                                                        })
                                                                    }} />
                                                                </Grid>
                                                            </Grid>
                                                        </Grid>
                                                    </Box>
                                                </Grid>

                                                <Grid item xs>
                                                    <Box bgcolor="#1F2933" component={Paper} sx={{ width: "calc(100% - 80px)", padding: "10px 40px" }}>
                                                        <Grid container spacing={"40px"} alignItems="center">
                                                            <Grid item xs="auto">
                                                                <Box bgcolor={"#25ee32"} sx={{ display: "flex", alignItems: "center", justifyContent: "center" }} width="100px" height="100px" borderRadius="250px" margin="25px 0">
                                                                    <CheckIcon sx={{ color: "white", fontSize: "75px" }} fontSize="large" />
                                                                </Box>
                                                            </Grid>
                                                            <Grid item>
                                                                <Typography color="#B2B3B7" fontSize="23px" fontWeight="400" textAlign="center">DMS is ready to enroll Devices</Typography>
                                                            </Grid>
                                                        </Grid>
                                                    </Box>
                                                </Grid>
                                            </Grid>

                                            <Grid item xs container>
                                                <Grid container spacing="40px">
                                                    <Grid item xs="auto" display="flex" alignItems="center" justifyContent="center" flexDirection="column">
                                                        <img src={process.env.PUBLIC_URL + "/assets/Chip.png"} height="175px" />
                                                    </Grid>
                                                    <ArrowSeparator
                                                        topLabelStep={1}
                                                        topLabelText="Requesting Certificate"
                                                        topLabelLoading={enrollmentProcesState.step === 0}
                                                        topLabelLoaded={enrollmentProcesState.step >= 1}
                                                        topUseWarn={false}
                                                        bottomLabelStep={4}
                                                        bottomLabelText="Transferring Certificate to Device"
                                                        bottomLabelLoading={enrollmentProcesState.step === 3 && enrollmentProcesState.authorizedCertificateTransfer === false}
                                                        bottomLabelLoaded={enrollmentProcesState.step >= 4}
                                                        bottomUseWarn={true}
                                                    />

                                                    <Grid item xs="auto" display="flex" alignItems="center" justifyContent="center" flexDirection="column">
                                                        <img src={process.env.PUBLIC_URL + "/assets/DMS.png"} height="175px" />
                                                    </Grid>

                                                    <ArrowSeparator
                                                        topLabelStep={2}
                                                        topLabelText="Generating Identity"
                                                        topLabelLoading={enrollmentProcesState.step === 1 && enrollmentProcesState.authorizedEnrollment === false}
                                                        topLabelLoaded={enrollmentProcesState.step >= 2}
                                                        topUseWarn={true}
                                                        bottomLabelStep={3}
                                                        bottomLabelText="Receiving Identity from PKI"
                                                        bottomLabelLoading={enrollmentProcesState.step === 2}
                                                        bottomLabelLoaded={enrollmentProcesState.step >= 3}
                                                        bottomUseWarn={false}
                                                    />
                                                    <Grid item xs="auto" display="flex" alignItems="center" justifyContent="center" flexDirection="column">
                                                        <img src={process.env.PUBLIC_URL + "/assets/SimpleIkerlan.svg"} width="175px" />
                                                    </Grid>
                                                </Grid>
                                            </Grid>

                                            <Grid item xs="auto" container spacing={4}>
                                                <Grid item xs container>
                                                    <Box bgcolor="#1F2933" component={Paper} padding="20px" flex="1">
                                                        {
                                                            enrollmentProcesState.step > 0
                                                                ? (
                                                                    <Grid container spacing={2}>
                                                                        <Grid item xs={6}>
                                                                            <Typography color="#B2B3B7" fontSize="15px" fontWeight="400">Requesting Date</Typography>
                                                                            <Typography color="#DEE2E7" fontSize="18px" fontWeight="400">{moment(enrollmentProcesState.requestingDate).format("DD/MM/YYYY HH:mm:ss")}</Typography>
                                                                        </Grid>
                                                                        <Grid item xs={6}>
                                                                            <Typography color="#B2B3B7" fontSize="15px" fontWeight="400">Device Model</Typography>
                                                                            <Typography color="#DEE2E7" fontSize="18px" fontWeight="400">{enrollmentProcesState.deviceModel}</Typography>
                                                                        </Grid>
                                                                        <Grid item xs={6}>
                                                                            <Typography color="#B2B3B7" fontSize="13px" fontWeight="400">Device ID</Typography>
                                                                            <Typography color="#DEE2E7" fontSize="18px" fontWeight="400">{enrollmentProcesState.deviceID}</Typography>
                                                                        </Grid>
                                                                        <Grid item xs={6}>
                                                                            <Typography color="#B2B3B7" fontSize="13px" fontWeight="400">Slot</Typography>
                                                                            <Typography color="#DEE2E7" fontSize="18px" fontWeight="400">{enrollmentProcesState.deviceSlot}</Typography>
                                                                        </Grid>
                                                                        <Grid item xs={12}>
                                                                            <Typography color="#B2B3B7" fontSize="15px" fontWeight="400">Certificate Signing Request</Typography>
                                                                            <Typography color="#DEE2E7" fontSize="18px" fontWeight="400">{enrollmentProcesState.certificateRequest}</Typography>
                                                                        </Grid>
                                                                        <Grid item xs={6}>
                                                                            <Typography color="#B2B3B7" fontSize="15px" fontWeight="400">Actions</Typography>
                                                                            <Grid container spacing={2}>
                                                                                <Grid item>
                                                                                    {
                                                                                        dmsState.autoEnrollment
                                                                                            ? (
                                                                                                <Typography color="#DEE2E7" fontStyle="italic" fontSize="23px" fontWeight="400">Automatic enrollment authorization is enabled</Typography>
                                                                                            )
                                                                                            : (
                                                                                                <Button variant="contained" onClick={() => {
                                                                                                    dispatch({
                                                                                                        type: ActionType.WS_SEND_MESSAGE,
                                                                                                        value: {
                                                                                                            type: "AUTH_ENROLL",
                                                                                                            message: {},
                                                                                                            time: Date.now()
                                                                                                        }
                                                                                                    })
                                                                                                }}>Authorize</Button>
                                                                                            )
                                                                                    }
                                                                                </Grid>
                                                                            </Grid>
                                                                        </Grid>
                                                                    </Grid>
                                                                )
                                                                : (
                                                                    <Typography color="#B2B3B7" fontSize="23px" fontWeight="400">No enrollment process underway</Typography>
                                                                )
                                                        }
                                                    </Box>
                                                </Grid>
                                                <Grid item xs container>
                                                    <Box bgcolor="#1F2933" component={Paper} padding="20px" flex="1">
                                                        {
                                                            enrollmentProcesState.step > 2
                                                                ? (
                                                                    <Grid container spacing={2}>
                                                                        <Grid item xs={6}>
                                                                            <Typography color="#B2B3B7" fontSize="15px" fontWeight="400">Certificate Serial Number</Typography>
                                                                            <Typography color="#DEE2E7" fontSize="18px" fontWeight="400">{enrollmentProcesState.serialNumber}</Typography>
                                                                        </Grid>
                                                                        <Grid item xs={6}>
                                                                            <Typography color="#B2B3B7" fontSize="15px" fontWeight="400">Expiration Date</Typography>
                                                                            <Typography color="#DEE2E7" fontSize="18px" fontWeight="400">{moment(enrollmentProcesState.expirationDate).format("DD/MM/YYYY HH:mm:ss")}</Typography>
                                                                        </Grid>
                                                                        <Grid item xs={12}>
                                                                            <Typography color="#B2B3B7" fontSize="15px" fontWeight="400">Certificate</Typography>
                                                                            <Typography color="#DEE2E7" fontSize="18px" fontWeight="400">{enrollmentProcesState.certificate}</Typography>
                                                                        </Grid>
                                                                        <Grid item xs={6}>
                                                                            <Typography color="#B2B3B7" fontSize="15px" fontWeight="400">Actions</Typography>
                                                                            <Grid container spacing={2}>
                                                                                <Grid item>
                                                                                    {
                                                                                        dmsState.autoCertificateTransfer
                                                                                            ? (
                                                                                                <Typography color="#DEE2E7" fontStyle="italic" fontSize="18px" fontWeight="400">Automatic transfer is enabled</Typography>
                                                                                            )
                                                                                            : (
                                                                                                <Button variant="contained" onClick={() => {
                                                                                                    dispatch({
                                                                                                        type: ActionType.WS_SEND_MESSAGE,
                                                                                                        value: {
                                                                                                            type: "AUTH_TRANSFER",
                                                                                                            message: {},
                                                                                                            time: Date.now()
                                                                                                        }
                                                                                                    })
                                                                                                }}>Transfer Certificate To Device</Button>
                                                                                            )
                                                                                    }
                                                                                </Grid>
                                                                            </Grid>
                                                                        </Grid>
                                                                    </Grid>
                                                                )
                                                                : (
                                                                    enrollmentProcesState.step === 0
                                                                        ? (
                                                                            <Typography color="#B2B3B7" fontSize="23px" fontWeight="400">No enrollment process underway </Typography>
                                                                        )
                                                                        : (
                                                                            <Typography color="#B2B3B7" fontSize="23px" fontWeight="400">Awaiting PKI Response ...</Typography>
                                                                        )
                                                                )
                                                        }
                                                    </Box>
                                                </Grid>
                                            </Grid>
                                        </Grid>
                                    )
                            }

                        </Grid>
                    </Box>
                </Grid>
                {
                    showSidebar && (
                        <Grid item xs={3} bgcolor="#1E1E1E" height="100%" overflow="auto">
                            <Box component={Paper} elevation={0} bgcolor="#1E1E1E" flex={1} height="100%" borderRadius={0} padding="20px">
                                <Grid container spacing={2}>
                                    <Grid item xs={12} container spacing={1}>
                                        <Grid item xs={12}>
                                            <Typography color="#B2B3B7" fontSize="23px" fontWeight="400">Web Socket Messages</Typography>
                                        </Grid>
                                        <Grid item xs={12}>
                                            <Button size="medium" variant="outlined" startIcon={<DeleteOutlineOutlinedIcon />} onClick={() => {
                                                dispatch({
                                                    type: ActionType.WS_CLEAR_MESSAGES
                                                })
                                            }}>
                                                Clear Messages
                                            </Button>
                                        </Grid>
                                        <Grid item xs={12} container spacing="20px">
                                            {
                                                websocketMessages.map((wsMessage, idx) => (
                                                    <Grid item xs={12} container key={idx}>
                                                        {
                                                            <>
                                                                <Grid item xs={12} container spacing={1}>
                                                                    <Grid item>
                                                                        <Typography fontSize="10px" fontWeight="500">{wsMessage.origin}</Typography>
                                                                    </Grid>
                                                                    <Grid item>
                                                                        <Typography fontSize="10px" fontWeight="500">{moment(wsMessage.timestamp).format("DD/MM/YYYY HH:mm:ss")}</Typography>
                                                                    </Grid>
                                                                </Grid>
                                                                <Grid item xs={12}>
                                                                    <Typography fontSize="10px">{JSON.stringify(wsMessage.message)}</Typography>
                                                                </Grid>
                                                            </>
                                                        }
                                                    </Grid>
                                                ))
                                            }
                                        </Grid>
                                    </Grid>
                                </Grid>
                            </Box>

                        </Grid>
                    )
                }
            </Grid>
        </ThemeProvider>
    )
}

export default App
