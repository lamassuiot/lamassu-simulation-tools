import { MiddlewareAPI } from "redux"
import { TypeSocket } from "typesocket"

import { ActionType } from "./actionTypes"
import { MessageModel } from "./models"

export const socketMiddleware = (url: string) => {
    return (store: MiddlewareAPI<any, any>) => {
        const socket = new TypeSocket<MessageModel >(url)

        // We dispatch the actions for further handling here:
        socket.on("connected", () => store.dispatch({ type: ActionType.WS_CONNECTED }))
        socket.on("disconnected", () => store.dispatch({ type: ActionType.WS_DISCONNECTED }))
        socket.on("message", (message) => store.dispatch({ type: ActionType.WS_MESSAGE, value: message }))
        socket.connect()

        let pendingMessages: Array<any> = []
        let intervalID:any

        return (next: (action: any) => void) => (action: any) => {
            // We're acting on an action with type of WS_SEND_MESSAGE.
            // Don't forget to check if the socket is in readyState == 1.
            // Other readyStates may result in an exception being thrown.
            if (action.type && action.type === ActionType.WS_SEND_MESSAGE) {
                if (socket.readyState !== 1) {
                    pendingMessages.push(action.value)
                    if (intervalID === undefined) {
                        intervalID = setInterval(() => {
                            if (socket.readyState === 1) {
                                pendingMessages.forEach(message => {
                                    socket.send(message)
                                })
                                pendingMessages = []
                                clearInterval(intervalID)
                                intervalID = undefined
                            }
                        }, 1000)
                    }
                } else {
                    socket.send(action.value)
                }
            }

            if (action.type && action.type === ActionType.WS_RECONNECT && socket.readyState === 3) {
                socket.connect()
            }

            return next(action)
        }
    }
}
