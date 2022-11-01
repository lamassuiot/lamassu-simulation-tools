import React from "react"
import { Box, Grid, keyframes, Typography } from "@mui/material"

interface Props {
    topLabelStep: number
    topLabelText: string
    topLabelLoading: boolean
    topLabelLoaded: boolean
    topUseWarn: boolean

    bottomLabelStep: number
    bottomLabelText: string
    bottomLabelLoading: boolean
    bottomLabelLoaded: boolean
    bottomUseWarn: boolean

}

export const ArrowSeparator: React.FC<Props> = ({ topLabelStep, topLabelText, topLabelLoading, topLabelLoaded, topUseWarn, bottomLabelStep, bottomLabelText, bottomLabelLoading, bottomLabelLoaded, bottomUseWarn }) => {
    const blinker = keyframes`
    50% {
        opacity: 0;
      }
    `

    const green = "#25ee32"
    const white = "#B2B3B7"
    const orange = "#FFA32A"

    const topColor = topLabelLoaded ? green : (topLabelLoading ? (topUseWarn ? orange : white) : white)
    const bottomColor = bottomLabelLoaded ? green : (bottomLabelLoading ? (bottomUseWarn ? orange : white) : white)

    return (
        <Grid item xs container alignItems="center" justifyContent="center" flexDirection="column" spacing={4}>
            <Grid item container alignItems="center" width="100%" spacing={2}>
                <Grid item container justifyContent="center" >
                    <Box color={topColor} borderRadius="40px" border={`1px solid ${topColor}`} display="flex" alignItems="center" justifyContent="center" width="30px" height="30px" marginRight="10px" fontSize="18px"> {topLabelStep} </Box>
                    <Typography color={topColor} fontSize="23px" fontWeight="400">{topLabelText}</Typography>
                </Grid>
                <Grid item container alignItems="center" justifyContent="flex-end" sx={{
                    animation: topLabelLoading ? `${blinker} 1.5s linear infinite` : "none"
                }}>
                    <Box bgcolor={topColor} width="100%" height="7px" />
                    <Box border={`solid ${topColor}`} sx={{ borderWidth: "0 8px 8px 0", transform: "rotate(-45 deg)", "-webkit-transform": "rotate(-45deg)", position: "relative", right: "0px", top: "-12px" }} display="inline-block" padding="5px" />
                </Grid>
            </Grid>

            <Grid item container alignItems="center" width="100%" spacing={1}>
                <Grid item container alignItems="center" justifyContent="flex-start" sx={{
                    animation: bottomLabelLoading ? `${blinker} 1.5s linear infinite` : "none"
                }}>
                    <Box bgcolor={bottomColor} width="100%" height="7px" />
                    <Box border={`solid ${bottomColor}`} sx={{ borderWidth: "0 8px 8px 0", transform: "rotate(135 deg)", "-webkit-transform": "rotate(135deg)", position: "relative", right: "7px", top: "-12px" }} display="inline-block" padding="5px" />
                </Grid>
                <Grid item container justifyContent="center" >
                    <Box color={bottomColor} borderRadius="40px" border={`1px solid ${bottomColor}`} display="flex" alignItems="center" justifyContent="center" width="30px" height="30px" marginRight="10px" fontSize="18px"> {bottomLabelStep} </Box>
                    <Typography color={bottomColor} fontSize="23px" fontWeight="400">{bottomLabelText}</Typography>
                </Grid>
            </Grid>
        </Grid>
    )
}
