import * as React from "react"
import {HashRouter, Route, Routes,} from "react-router-dom"
import Start from "./Start"
import Layout from "./Layout"
import SignOn from "./SignOn";
import SignUp from "./SignUp";
import RequireUser from "./RequireUser";
import RequireNoUser from "./RequireNoUser";


export default function App() {
	return (
		<HashRouter>
			<Routes>
				<Route path="/" element={<Layout/>}>
					<Route path="/" element={<RequireUser><Start/></RequireUser>}/>
					<Route path="/signon" element={<RequireNoUser><SignOn/></RequireNoUser>}/>
					<Route path="/signup" element={<RequireNoUser><SignUp/></RequireNoUser>}/>
				</Route>
			</Routes>
		</HashRouter>
	)
}
