import { all, put, takeEvery } from "redux-saga/effects"
import { MessageModel } from "./models"
import { ActionType as ActionTypeDMS } from "./features/dms/actionTypes"
import { ActionType as ActionTypeEnrollProcess } from "./features/enrollProcesor/actionTypes"
import { ActionType as ActionTypeWS } from "./features/websocket/actionTypes"

function * message (action: any) {
    console.log(action)

    const msg: MessageModel = action.value as MessageModel

    // Now we can act on incoming messages
    switch (msg.type) {
    case ActionTypeEnrollProcess.ENROLLING_PROCESS_UPDATE:
        yield put({ type: ActionTypeEnrollProcess.ENROLLING_PROCESS_UPDATE, value: msg })
        break

    case ActionTypeDMS.DMS_UPDATE:
        yield put({ type: ActionTypeDMS.DMS_UPDATE, value: msg })
        break

    case ActionTypeDMS.ENROLLED_IDENTITES_UPDATE:
        yield put({ type: ActionTypeDMS.ENROLLED_IDENTITES_UPDATE, value: msg })
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
