import { Tabs } from "antd";
import General from "./General";
import AccountsPool from "./AccountsPool";

const tabItems = [
  {
    key: "general",
    label: "General",
    children: <General />,
  },
  {
    key: "accounts-pool",
    label: "Worker Accounts Pool",
    children: <AccountsPool />,
  },
];

const PikPak = () => {
  return (
    <>
      <Tabs items={tabItems}></Tabs>
    </>
  );
};

export default PikPak;
