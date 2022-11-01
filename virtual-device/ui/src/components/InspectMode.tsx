/* eslint-disable */
import React, { useState, useEffect } from "react"
import { Box, Button, ButtonGroup, createTheme, Grid, IconButton, keyframes, Paper, Slider, ThemeProvider, Typography } from "@mui/material"
import CachedIcon from "@mui/icons-material/Cached"
import DeleteOutlineOutlinedIcon from "@mui/icons-material/DeleteOutlineOutlined"
import moment from "moment"
import { useAppSelector } from "ducks/hooks"
import * as websocketSelector from "ducks/features/websocket/reducer"
import * as deviceManagerSelector from "ducks/features/deviceManager/reducer"
import { useDispatch } from "react-redux"
import { ActionType } from "ducks/features/websocket/actionTypes"

const InspectMode= () => {
    const dispatch = useDispatch()

    const websocketMessages = useAppSelector((state: any) => websocketSelector.getMessages(state))
    console.log(websocketMessages)

    const [provisioningFlow, setProvisioningFlow] = useState("device")
    const supportedProvisioningFlows = [
        "device",
        "dms"
    ]

    return (

        <Grid item xs={3} bgcolor="#1E1E1E" height="100%" overflow="auto">
            <Box component={Paper} elevation={0} bgcolor="#1E1E1E" flex={1} height="100%" borderRadius={0} padding="25px">
                <Grid container spacing={2}>
                    <Grid item xs={12} container spacing={1}>
                        <Grid item xs={12}>
                            <Typography color="#B2B3B7" fontSize="25px" fontWeight="400">Provisioning Flow Mode</Typography>
                        </Grid>
                        <Grid item xs={12}>
                            <ButtonGroup size="medium" aria-label="medium button group">
                                {
                                    supportedProvisioningFlows.map((flowMode, idx) => (
                                        <Button key={idx} variant={provisioningFlow === flowMode ? "contained" : "outlined"} onClick={() => {
                                            setProvisioningFlow(flowMode)
                                        }}>Initated by {flowMode}</Button>
                                    ))
                                }
                            </ButtonGroup>
                        </Grid>
                    </Grid>
                    <Grid item xs={12}>
                        <Typography color="#B2B3B7" fontSize="25px" fontWeight="400">Device Telemetry Data Rate (seconds)</Typography>
                        <Slider
                            size="medium"
                            min={1}
                            defaultValue={10}
                            max={30}
                            aria-label="medium"
                            valueLabelDisplay="auto"
                            onChangeCommitted={(ev, newValue) => {
                                dispatch({
                                    type: ActionType.WS_SEND_MESSAGE,
                                    value: {
                                        type: "CHANGE_TELEMETRY_DATA_RATE",
                                        message: {
                                            new_rate: newValue
                                        },
                                        time: Date.now()
                                    }
                                })
                            }}
                        />
                    </Grid>
                    <Grid item xs={12} container spacing={1}>
                        <Grid item xs={12}>
                            <Typography color="#B2B3B7" fontSize="25px" fontWeight="400">Web Socket Messages</Typography>
                        </Grid>
                        <Grid item xs={12}>
                            <Button sx={{ height: "50px", fontSize: "30px" }} variant="outlined" startIcon={<DeleteOutlineOutlinedIcon />} onClick={() => {
                                dispatch({
                                    type: ActionType.WS_CLEAR_MESSAGES
                                })
                            }}>
                                Clear Messages
                            </Button>
                        </Grid>
                        <Grid item xs={12} container spacing="25px">
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

export default InspectMode
