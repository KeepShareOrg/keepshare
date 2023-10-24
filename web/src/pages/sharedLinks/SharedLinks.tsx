import TableFilter from "@/components/sharedLinks/TableFilter";
import SharedLinkTable from "../../components/sharedLinks/SharedLInkTable";
import { querySharedLinks } from "@/api/link";
import useStore from "@/store";
import { useEffect, useState } from "react";
import { shimSharedLinksTableData } from "@/util";
import { LinkFormatType } from "@/constant";
import { message } from "antd";

interface PaginationInfo {
  search: string;
  pageIndex: number;
  pageSize: number;
}
const Content = () => {
  const [
    setIsLoading,
    setTableData,
    setTotalSharedNum,
    totalSharedLinks,
    setTotalSharedLinks,
  ] = useStore((state) => [
    state.setIsLoading,
    state.setTableData,
    state.setTotalSharedNum,
    state.totalSharedLinks,
    state.setTotalSharedLinks,
  ]);

  const [pageInfo, setPageInfo] = useState<PaginationInfo>({
    search: "",
    pageIndex: 1,
    pageSize: 10,
  });
  const { search, pageIndex, pageSize } = pageInfo;

  const { data, error, isLoading, mutate } = querySharedLinks({
    search,
    limit: pageSize,
    pageIndex: pageIndex,
  });

  error && message.error("query shared links error!");

  useEffect(() => {
    const tableData = data?.data?.list;
    const totalSharedNum = data?.data?.total;
    if (!isLoading && Array.isArray(tableData)) {
      const shimTableData = tableData.map((v) =>
        shimSharedLinksTableData({ ...v }),
      );
      setTableData(shimTableData);

      const localTotalSharedLinksKeys = totalSharedLinks.map((v) => v.auto_id);
      const newTotalSharedLinks = shimTableData.filter(
        (v) => !localTotalSharedLinksKeys.includes(v.auto_id),
      );
      setTotalSharedLinks([...totalSharedLinks, ...newTotalSharedLinks]);

      totalSharedNum && setTotalSharedNum(totalSharedNum);
    }
    setIsLoading(isLoading);
  }, [isLoading, pageInfo, data]);

  const handlePageChange = (pageIndex: number, pageSize: number) => {
    setPageInfo({ search, pageIndex, pageSize });
  };

  const handleSearch = (search: string) => {
    setPageInfo({ search, pageIndex: 0, pageSize });
  };

  const [formatType, setFormatType] = useState<LinkFormatType>(
    LinkFormatType.TEXT,
  );
  const handleFormat = (formatType: LinkFormatType) =>
    setFormatType(formatType);

  return (
    <>
      <TableFilter handleFormat={handleFormat} handleSearch={handleSearch} />
      <SharedLinkTable
        refresh={mutate}
        formatType={formatType}
        handlePageChange={handlePageChange}
      />
    </>
  );
};

export default Content;
