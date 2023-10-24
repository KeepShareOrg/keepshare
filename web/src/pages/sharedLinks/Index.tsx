import { Tabs } from "antd";
import SharedLinksBlacklist from "./SharedLinksBlacklist";
import SharedLinksList from "./SharedLinks";
import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";

const enum TabKey {
  SHARED_LINKS = "shared-links",
  LINK_BLACKLIST = "link-blacklist",
}
const tabItems = [
  {
    key: TabKey.SHARED_LINKS,
    label: "Shared Links",
    children: <SharedLinksList />,
  },
  {
    key: TabKey.LINK_BLACKLIST,
    label: "Link Blacklist",
    children: <SharedLinksBlacklist />,
  },
];

const SharedLinks = () => {
  const [activeKey, setActiveKey] = useState(TabKey.SHARED_LINKS);
  const [query, setQuery] = useSearchParams();
  useEffect(() => {
    const tabKey = query.get("tab") as TabKey;
    if ([TabKey.SHARED_LINKS, TabKey.LINK_BLACKLIST].includes(tabKey)) {
      setActiveKey(tabKey);
      setQuery("");
    }
  }, [query]);

  return (
    <Tabs
      activeKey={activeKey}
      items={tabItems}
      onTabClick={(k) => setActiveKey(k as TabKey)}
    ></Tabs>
  );
};

export default SharedLinks;
