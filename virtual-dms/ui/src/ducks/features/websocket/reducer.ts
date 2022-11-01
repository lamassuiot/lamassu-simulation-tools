import { RootState } from "ducks/reducers"
import { actions } from "ducks/actions"
import moment from "moment"

interface MessageState {
    origin: string,
    timestamp: Date
    message: any
}

export interface WebSocketState {
    state: string,
    messages: Array<MessageState>
}

const initialState = {
    state: "DISCONNECTED",
    messages: []
}

export const websocketReducer = (state = initialState, action: any) => {
    console.log(action)

    switch (action.type) {
    case actions.webSocketsActions.ActionType.WS_CLEAR_MESSAGES:
        return Object.assign({}, state, {
            messages: []
        })
    case actions.webSocketsActions.ActionType.WS_CONNECTED:
        return Object.assign({}, state, {
            state: "CONNECTED"
        })
    case actions.webSocketsActions.ActionType.WS_DISCONNECTED:
        return Object.assign({}, state, {
            state: "DISCONNECTED"
        })
    case actions.webSocketsActions.ActionType.WS_SEND_MESSAGE:
        return Object.assign({}, state, {
            messages: [{
                origin: "OUT",
                timestamp: moment(),
                message: action.value
            }, ...state.messages]
        })
    case actions.webSocketsActions.ActionType.WS_MESSAGE:
        return Object.assign({}, state, {
            messages: [{
                origin: "IN",
                timestamp: moment(),
                message: action.value
            }, ...state.messages]
        })
    default:
        break
    }
    return state
}

const getSelector = (state: RootState): WebSocketState => state.websocket

export const getWebsocketState = (state: RootState): string => {
    const reducer = getSelector(state)
    return reducer.state
}
export const getMessages = (state: RootState): Array<MessageState> => {
    const reducer = getSelector(state)
    return reducer.messages
}
