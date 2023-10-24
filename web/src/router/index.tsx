import { createBrowserRouter, redirect } from "react-router-dom";
import type { LoaderFunctionArgs, RouteObject } from "react-router-dom";
import useStore from "@/store";
import { getTokenInfo } from "@/util";
import { Suspense, lazy } from "react";
// import ErrorPage from "@/pages/ErrorPage";
import PageLoading from "@/components/common/PageLoading";
const Home = lazy(() => import("@/pages/home/Home"));
const Login = lazy(() => import("@/pages/login/Login"));
const SignUp = lazy(() => import("@/pages/signUp/SignUp"));
const ResetPassword = lazy(() => import("@/pages/resetPassword/ResetPassword"));
const AutoShare = lazy(() => import("@/pages/autoShare/AutoShare"));
const LinkToShare = lazy(() => import("@/pages/linkToShare/LinkToShare"));
const SharedLinks = lazy(() => import("@/pages/sharedLinks/Index"));
const Settings = lazy(() => import("@/pages/settings/Settings"));
const PikPak = lazy(() => import("@/pages/management/pikpak/PikPak"));
const Mega = lazy(() => import("@/pages/management/mega/Mega"));
const SharedStatus = lazy(() => import("@/pages/sharedStatus/SharedStatus"));
const Donation = lazy(() => import("@/pages/donation/Donation"));

export const enum RoutePaths {
  Home = "/console/",
  Login = "/console/login",
  SignUp = "/console/sign-up",
  ResetPassword = "/console/reset-password",
  AutoShare = "/console/auto-share",
  LinkToShare = "/console/link-to-share",
  SharedLinks = "/console/shared-links",
  Settings = "/console/settings",
  PikPak = "/console/management/pikpak",
  Mega = "/console/management/mega",
  SharedStatus = "/console/shared/status",
  Donation = "/console/donation",
}

// Read some data from persistent storage into memory
const appInit = () => {
  const tokenInfo = getTokenInfo();
  // During initialization, if there is an accessToken in localStorage it is temporarily considered to be logged in
  tokenInfo.accessToken &&
    useStore.getState().signIn({
      accessToken: tokenInfo.accessToken,
      refreshToken: tokenInfo.refreshToken,
    });
};
appInit();

const authLoader = ({ request }: LoaderFunctionArgs) => {
  const isLogin = useStore.getState().isLogin;
  if (isLogin) {
    return null;
  }
  const params = new URLSearchParams();
  params.set("from", new URL(request.url).pathname);
  return redirect(`${RoutePaths.Login}?${params.toString()}`);
};

const suspenseWrapper = (
  routes: RouteObject[],
  LoadingAnimation: React.ReactNode,
) => {
  const newRoutes: RouteObject[] = [];
  routes.forEach((r) => {
    if (r.element) {
      r.element = <Suspense fallback={LoadingAnimation}>{r.element}</Suspense>;
    }
    if (r.children && r.children.length > 0) {
      r.children = suspenseWrapper(r.children, LoadingAnimation);
    }
    newRoutes.push(r);
  });

  return newRoutes;
};

const router = createBrowserRouter(
  suspenseWrapper(
    [
      {
        path: RoutePaths.Home,
        element: <Home />,
        loader: authLoader,
        // errorElement: <ErrorPage />,
        children: [
          {
            path: RoutePaths.AutoShare,
            element: <AutoShare />,
          },
          {
            path: RoutePaths.LinkToShare,
            element: <LinkToShare />,
          },
          {
            path: RoutePaths.SharedLinks,
            element: <SharedLinks />,
          },
          {
            path: RoutePaths.Settings,
            element: <Settings />,
          },
          {
            path: RoutePaths.PikPak,
            element: <PikPak />,
          },
          {
            path: RoutePaths.Mega,
            element: <Mega />,
          },
        ],
      },
      {
        path: RoutePaths.Login,
        element: <Login />,
      },
      {
        path: RoutePaths.SignUp,
        element: <SignUp />,
      },
      {
        path: RoutePaths.ResetPassword,
        element: <ResetPassword />,
      },
      {
        path: RoutePaths.SharedStatus,
        element: <SharedStatus />,
      },
      {
        path: RoutePaths.Donation,
        element: <Donation />,
      },
    ],
    <PageLoading />,
  ),
);

export default router;
