import * as React from "react"
import {BrowserRouter, HashRouter, Route, Routes,} from "react-router-dom"
import Start from "./Start"
import Layout from "./Layout"
import SignOn from "./SignOn";
import SignUp from "./SignUp";
import RequireUser from "./RequireUser";


export default function App() {
	return (
		<HashRouter>
			<Routes>
				<Route path="/" element={<Layout/>}>
					<Route path="/" element={<RequireUser><Start/></RequireUser>}/>
					<Route path="/signon" element={<SignOn/>}/>
					<Route path="/signup" element={<SignUp/>}/>
				</Route>
			</Routes>
		</HashRouter>
	)
}
