import useStore from "@/store";
import { useEffect, useState } from "react";
import {
  Typography,
  Space,
  Form,
  Input,
  Select,
  Divider,
  Button,
  FormProps,
  message,
} from "antd";
import DonationPng from "@/assets/images/donation-png.png";
import LogoWithText from "@/assets/images/logo-with-text.png";
import { ArrowRightOutlined } from "@ant-design/icons";
import { useSearchParams } from "react-router-dom";
import { donateRedeemCode } from "@/api/pikpak";

const { Title, Text, Paragraph } = Typography;
const { TextArea } = Input;

const ItemLabel = (text: string) => (
  <Text style={{ width: "112px", textAlign: "right" }}>{text}</Text>
);

interface FieldType {
  nickname?: string;
  channelID?: string;
  drive: string;
  redeemCodes: string;
}

const Donation = () => {
  const [formData] = Form.useForm<FieldType>();
  const [params] = useSearchParams();
  const setThemeMode = useStore((state) => state.setThemeMode);

  const [existInitialChannelId, setExistInitialChannelId] = useState(false);
  useEffect(() => {
    setThemeMode("light");

    const channelID = params.get("channel") || "";
    if (channelID) {
      formData.setFieldsValue({
        drive: "pikpak",
        channelID: params.get("channel") || "",
      });
      setExistInitialChannelId(true);
    }
  }, []);

  const handleFinish: FormProps<FieldType>["onFinish"] = async (values) => {
    console.log("handle finish: ", values);
    const { error } = await donateRedeemCode({
      nickname: values.nickname,
      channel_id: values.channelID || "",
      drive: values.drive,
      redeem_codes: values.redeemCodes.split(",").map((item) => item.trim()),
    });
    if (error != null) {
      message.error(error.error);
    } else {
      message.success("Thank you for your donation!");
    }
  };

  return (
    <>
      <Space>
        <img
          src={LogoWithText}
          style={{ height: "64px", marginTop: "12px", marginLeft: "24px" }}
        />
      </Space>
      <Paragraph style={{ maxWidth: "1200px", marginInline: "auto" }}>
        <Space>
          <Space direction="vertical">
            <Title level={3}>KeepShare Premium code donation</Title>
            <Text>
              Remote uploading and sharing preparations will begin immediately
              after submission. You can post these created shared links, or you
              can also pre-create Auto-Sharing links so that peoples can get the
              shared files as soon as possible when accessing the keep sharing
              link...
            </Text>
            <Button type="link" style={{ marginTop: "24px", paddingInline: 0 }}>
              Go to Github to view donation records
              <ArrowRightOutlined />
            </Button>
          </Space>
          <img src={DonationPng} style={{ height: "200px" }} />
        </Space>

        <Divider />
        <Form
          form={formData}
          style={{ maxWidth: "600px" }}
          onFinish={handleFinish}
        >
          <Form.Item<FieldType> label={ItemLabel("Your nickname")}>
            <Form.Item<FieldType> name="nickname" noStyle>
              <Input />
            </Form.Item>
            <Text type="secondary">
              If you want to remain anonymous, you can leave the content blank
            </Text>
          </Form.Item>
          <Form.Item<FieldType>
            label={ItemLabel("Channel ID")}
            name="channelID"
          >
            <Input disabled={existInitialChannelId} />
          </Form.Item>
          <Form.Item<FieldType>
            label={ItemLabel("The cloud drive")}
            name="drive"
          >
            <Select options={[{ value: "PikPak", label: "PikPak" }]} />
          </Form.Item>
          <Form.Item<FieldType>
            label={ItemLabel("premium code")}
            name="redeemCodes"
          >
            <TextArea
              rows={5}
              placeholder="AutoSize height based on content lines"
            ></TextArea>
          </Form.Item>
          <Form.Item<FieldType> label={ItemLabel("")} colon={false}>
            <Button type="primary" htmlType="submit">
              Submit
            </Button>
          </Form.Item>
        </Form>
      </Paragraph>
    </>
  );
};

export default Donation;
