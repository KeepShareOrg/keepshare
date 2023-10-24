import loading from "@/assets/images/keepshare-loading.png";

const PageLoading = () => {
  return (
    <div
      style={{
        display: "flex",
        width: "100%",
        height: "100%",
        justifyContent: "center",
        alignItems: "center",
        flexDirection: "column",
        color: "#5B67EA",
      }}
    >
      <img src={loading} alt="loading" width={100} />
      <div style={{ fontSize: "20px" }}> Loading.... </div>
    </div>
  );
};

export default PageLoading;
