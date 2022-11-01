import { RootState } from "ducks/reducers"
import { actions } from "ducks/actions"

export interface EnrollProcesorState {
    step: number,
    requestingDate: Date,
    deviceModel: string,
    deviceID: string,
    deviceSlot: string,
    certificateRequest: string,
    authorizedEnrollment: boolean,
    certificate: string,
    serialNumber: string,
    expirationDate: Date,
    authorizedCertificateTransfer: boolean,
}

const initialState = {
    step: 0,
    requestingDate: undefined,
    deviceModel: "",
    deviceID: "",
    deviceSlot: "",
    certificateRequest: "",
    authorizedEnrollment: false,
    certificate: "",
    serialNumber: "",
    expirationDate: undefined,
    authorizedCertificateTransfer: false
}

export const enrollProcesorReducer = (state = initialState, action: any) => {
    console.log(action)

    switch (action.type) {
    case actions.enrollProcesorActions.ActionType.ENROLLING_PROCESS_UPDATE: {
        const statusStep = action.value.message.status.split("_")
        console.log(statusStep)
        return Object.assign({}, state, {
            step: parseInt(statusStep[1]),
            requestingDate: action.value.message.requesting_date,
            deviceModel: action.value.message.device_model,
            deviceID: action.value.message.device_id,
            deviceSlot: action.value.message.device_slot,
            certificateRequest: action.value.message.certificate_request,
            authorizedEnrollment: action.value.message.authorized_enrollment,
            certificate: action.value.message.certificate,
            serialNumber: action.value.message.serial_number,
            expirationDate: action.value.message.expiration_date,
            authorizedCertificateTransfer: action.value.message.authorized_certificate_transfer
        })
    }
    }
    return state
}

const getSelector = (state: RootState): EnrollProcesorState => state.enrollProcesor

export const getState = (state: RootState): EnrollProcesorState => {
    const reducer = getSelector(state)
    return reducer
}
