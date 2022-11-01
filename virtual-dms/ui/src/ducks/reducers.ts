import { combineReducers } from "redux"
import { dmsReducer, DMSState } from "./features/dms/reducer"
import { enrollProcesorReducer, EnrollProcesorState } from "./features/enrollProcesor/reducer"
import { websocketReducer, WebSocketState } from "./features/websocket/reducer"

export type RootState = {
  enrollProcesor: EnrollProcesorState,
  websocket: WebSocketState,
  dms: DMSState,
}

const reducers = combineReducers({
    enrollProcesor: enrollProcesorReducer,
    websocket: websocketReducer,
    dms: dmsReducer
})

export default reducers
