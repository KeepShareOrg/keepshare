import ReactDOM from "react-dom/client";
import GlobalStyle from "@/common/style/globalStyle";
import App from "./App";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <>
    <GlobalStyle />
    <App />
  </>,
);
