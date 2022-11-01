import { all, put, takeEvery } from "redux-saga/effects"
import { MessageModel } from "./features/websocket/models"
import { ActionType } from "./features/deviceManager/actionTypes"
import { ActionType as ActionTypeWS } from "./features/websocket/actionTypes"

function * message (action: any) {
    console.log(action)

    const msg: MessageModel = action.value as MessageModel

    // Now we can act on incoming messages
    switch (msg.type) {
    case ActionType.TELEMETRY_DATA:
        yield put({ type: ActionType.TELEMETRY_DATA, value: msg })
        break
    case ActionType.DEVICE_UPDATED:
        yield put({ type: ActionType.DEVICE_UPDATED, value: msg })
        break
    case ActionType.MQTT_LOG:
        yield put({ type: ActionType.MQTT_LOG, value: msg })
        break
    }
}
function * mySaga () {
    yield takeEvery(ActionTypeWS.WS_MESSAGE, message)
}

export default function * rootSaga () {
    yield all([
        mySaga()
    ])
}
