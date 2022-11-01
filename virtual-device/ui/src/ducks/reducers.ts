import { combineReducers } from "redux"
import { deviceManagerReducer, DeviceManagerState } from "./features/deviceManager/reducer"
import { websocketReducer, WebSocketState } from "./features/websocket/reducer"

export type RootState = {
  deviceManager: DeviceManagerState,
  websocket: WebSocketState

}

const reducers = combineReducers({
    deviceManager: deviceManagerReducer,
    websocket: websocketReducer
})

export default reducers
