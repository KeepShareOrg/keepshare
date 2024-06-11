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

    formData.setFieldsValue({ drive: "pikpak" });
    const channelID = params.get("channel") || "";
    if (channelID) {
      formData.setFieldsValue({
        channelID: params.get("channel") || "",
      });
      setExistInitialChannelId(true);
    }
  }, []);

  const handleFinish: FormProps<FieldType>["onFinish"] = async (values) => {
    console.log("handle finish: ", values);
    if (!values.redeemCodes?.trim()) {
      message.error("Please enter the redeem codes");
      return;
    }
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
              Although KeepShare is free, some cloud drive services have file
              storage limitations. Free accounts may not be sufficient to create
              shareable links for large files. To continue providing efficient
              and reliable services, we need some premium accounts or premium
              codes to support the advanced features of these cloud storage
              services.
            </Text>
            <Text>
              Every donation is a great support to our work. Thank you for
              helping us keep KeepShare free and open-source, providing
              convenience to users worldwide.
            </Text>
            <Button
              type="link"
              href="https://github.com/keepshareorg/keepshare"
              style={{ marginTop: "24px", paddingInline: 0 }}
            >
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
              <Input placeholder="Nickname" />
            </Form.Item>
            <Text type="secondary">
              If you want to remain anonymous, you can leave the content blank
            </Text>
          </Form.Item>
          <Form.Item<FieldType> label={ItemLabel("Channel ID")}>
            <Form.Item<FieldType> name="channelID" noStyle>
              <Input
                disabled={existInitialChannelId}
                placeholder="Channel ID"
              />
            </Form.Item>
            {existInitialChannelId && (
              <Text type="secondary">
                The donations are only allowed to be used in the current
                channel.
              </Text>
            )}
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
            <TextArea rows={5}></TextArea>
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
