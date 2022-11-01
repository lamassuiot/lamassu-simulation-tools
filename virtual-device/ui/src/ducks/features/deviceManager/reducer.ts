import { RootState } from "ducks/reducers"
import { actions } from "ducks/actions"
import moment from "moment"

export interface TelemetryDataState {
    battery_level: string
    humidity: string
    temperature: string
}

export interface SlotState {
    id: string,
    status: string,
    serialNumber: string,
    certificate: string,
    privateKey: string,
    issuingCA: string,
    expirationDate: Date
}

export interface DeviceState {
    status: string,
    serialNumber: string,
    model: string,
    slots: Array<SlotState>
    mqttConnected: boolean,
}

export interface MQTTLog {
    type: string,
    title: string,
    message: string,
    timestamp: Date
}

export interface DeviceManagerState {
    telemetryData: TelemetryDataState,
    device: DeviceState
    mqttLogs: Array<MQTTLog>
}

const initialState = {
    telemetryData: {
        battery_level: "-",
        humidity: "-",
        temperature: "-"
    },
    device: {
        status: "-",
        serialNumber: "-",
        model: "-",
        slots: []
    },
    mqttLogs: []
}

export const deviceManagerReducer = (state = initialState, action: any) => {
    console.log(action)

    switch (action.type) {
    case actions.deviceManagerActions.ActionType.DEVICE_UPDATED:
        return Object.assign({}, state, {
            device: {
                status: action.value.message.status,
                serialNumber: action.value.message.serial_number,
                model: action.value.message.model,
                mqttConnected: action.value.message.mqtt_connected,
                slots: action.value.message.slots.map((slot: any) => {
                    return {
                        id: slot.id,
                        status: slot.status,
                        serialNumber: slot.serial_number,
                        certificate: slot.certificate,
                        privateKey: slot.private_key,
                        issuingCA: slot.issuing_ca,
                        expirationDate: moment.unix(slot.expiration_date)
                    }
                })
            },
            telemetryData: {
                battery_level: action.value.message.telemetry_data.battery_level,
                humidity: action.value.message.telemetry_data.humidity,
                temperature: action.value.message.telemetry_data.temperature
            }
        })
    case actions.deviceManagerActions.ActionType.MQTT_LOG: {
        const logs = state.mqttLogs.slice(0, 20)
        console.log(logs, action.value)
        return Object.assign({}, state, {
            mqttLogs: [action.value.message, ...logs]
        })
    }
    default:
        break
    }
    return state
}

const getSelector = (state: RootState): DeviceManagerState => state.deviceManager

export const getTelemetryData = (state: RootState): TelemetryDataState => {
    const reducer = getSelector(state)
    return reducer.telemetryData
}

export const getDeviceState = (state: RootState): DeviceState => {
    const reducer = getSelector(state)
    return reducer.device
}

export const getMqttLogs = (state: RootState): Array<MQTTLog> => {
    const reducer = getSelector(state)
    return reducer.mqttLogs
}
