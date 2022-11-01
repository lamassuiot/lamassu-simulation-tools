/* eslint-disable @typescript-eslint/ban-types */
import createSagaMiddleware from "redux-saga"
import reducers from "./reducers"
import { createStore, applyMiddleware, compose } from "redux"
import sagas from "./sagas"
import { socketMiddleware } from "./features/websocket/socketMiddleware"

declare global {
  interface Window {
    __REDUX_DEVTOOLS_EXTENSION_COMPOSE__: Function
  }
}
const composeEnhancers =
  typeof window === "object" && window.__REDUX_DEVTOOLS_EXTENSION_COMPOSE__
      ? window.__REDUX_DEVTOOLS_EXTENSION_COMPOSE__({
      // Specify extension’s options like name, actionsDenylist, actionsCreators, serialize...
          serialize: true
      })
      : compose

const sagaMiddleware = createSagaMiddleware()

function configureStore () {
    const middleware = [socketMiddleware(`ws://${location.host}/ws`), sagaMiddleware]
    // const middleware = [socketMiddleware("ws://dev-lamassu.zpd.ikerlan.es:7002/ws"), sagaMiddleware]
    const enhancer = composeEnhancers(applyMiddleware(...middleware))
    return createStore(reducers, enhancer)
}

export const store = configureStore()
export type AppDispatch = typeof store.dispatch

sagaMiddleware.run(sagas as any)
