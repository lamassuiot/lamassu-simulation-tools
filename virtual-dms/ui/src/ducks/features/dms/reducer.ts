import { RootState } from "ducks/reducers"
import { actions } from "ducks/actions"

interface EnrolledIdentity {
    enrolled_timestamp: Date
    serial_number: string
    device_id: string
    device_slot: string
    issuing_ca: string
    issuing_duration: number
}

export interface DMSState {
    status: string,
    name: string,
    authorizedCAs: Array<string>,
    selectedCA: string,
    deviceSlot: string,
    autoEnrollment: boolean,
    autoCertificateTransfer: boolean,
    enrolledIdentities: Array<EnrolledIdentity>,
}

const initialState = {
    status: "EMPTY",
    name: "",
    authorizedCAs: [],
    selectedCA: "",
    autoEnrollment: false,
    autoCertificateTransfer: false,
    enrolledIdentities: []
}

export const dmsReducer = (state = initialState, action: any) => {
    console.log(action)

    switch (action.type) {
    case actions.dmsActions.ActionType.DMS_UPDATE:
        return Object.assign({}, state, {
            status: action.value.message.status,
            name: action.value.message.name,
            authorizedCAs: action.value.message.authorized_cas,
            selectedCA: action.value.message.selected_ca_for_enrollment,
            autoEnrollment: action.value.message.automatic_enrollment,
            autoCertificateTransfer: action.value.message.automatic_certificate_transfer
        })
    case actions.dmsActions.ActionType.ENROLLED_IDENTITES_UPDATE:
        return Object.assign({}, state, {
            enrolledIdentities: action.value.message.sort((a:EnrolledIdentity, b: EnrolledIdentity) => { return a.enrolled_timestamp > b.enrolled_timestamp ? -1 : 1 })
        })
    }
    return state
}

const getSelector = (state: RootState): DMSState => state.dms

export const getState = (state: RootState): DMSState => {
    const reducer = getSelector(state)
    return reducer
}
