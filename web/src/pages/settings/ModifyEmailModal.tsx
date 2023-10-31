import { useState } from 'react';
import { useNavigate } from "react-router-dom";
import { Modal, Form, message } from "antd";

import { changeAccountEmail } from '@/api/account';
import { calcPasswordHash } from '@/util';
import { RoutePaths } from "@/router";
import useStore from "@/store";

import { StyledForm, StyledInput, PasswordInput } from './style';

interface FieldType {
  email?: string;
  emailRepeat?: string;
  password?: string;
}

interface ModifyEmailModalProps {
  open: boolean;
  onClose: () => void;
}

function ModifyEmailModal(props: ModifyEmailModalProps) {
  const { open } = props;
  const [modifyEmailConfirm, setModifyEmailConfirm] = useState(false);
  const [isModifyEmailLoading, setIsModifyEmailLoading] = useState(false);
  const [form] = Form.useForm<FieldType>();
  const navigate = useNavigate();

  const signOut = useStore((state) => state.signOut);

  const handleModifyEmailModalClose = () => {
    props.onClose && props.onClose();
    form.resetFields();
  };

  const handleModifyEmailFormChange = async () => {
    try {
      await form.validateFields({
        validateOnly: true,
      })
      setModifyEmailConfirm(true);
    } catch {
      setModifyEmailConfirm(false);
    }
  };

  const handleModifyEmail = async ({ email, password }: FieldType) => {
    if (!email || !password) {
      console.error('can not find form field.');
      return;
    }

    setIsModifyEmailLoading(true);
    const result = await changeAccountEmail({
      new_email: email!,
      password_hash: calcPasswordHash(password!),
    });
    setIsModifyEmailLoading(false);

    if (result.error) {
      message.error(result.error.message || 'modify email fail');
      return;
    }

    message.success('modify email success!');
    signOut();
    // navigate to login
    navigate(RoutePaths.Login);
  };

  return (
    <Modal
      title="Modify Email"
      okText="Confirm"
      open={open}
      onCancel={handleModifyEmailModalClose}
      onOk={() => form.submit()}
      okButtonProps={{
        disabled: !modifyEmailConfirm && !isModifyEmailLoading,
        loading: isModifyEmailLoading,
      }}
    >
      <StyledForm
        form={form}
        layout="vertical"
        onFinish={(values) => handleModifyEmail(values as unknown as FieldType)}
        onValuesChange={handleModifyEmailFormChange}
        validateTrigger="onBlur"
        autoComplete="off"
      >
        <Form.Item<FieldType>
          label="New Email"
          name="email"
          rules={[
            { required: true, message: 'email is required' },
            { type: 'email', message: 'email address not valid' }
          ]}
        >
          <StyledInput placeholder="new email" />
        </Form.Item>
        <Form.Item<FieldType>
          label="New Email (Repeat)"
          name="emailRepeat"
          rules={[
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (value !== getFieldValue('email')) {
                  return Promise.reject(new Error('Email is inconsistent'));
                }
                return Promise.resolve();
              }
            }),
          ]}
        >
          <StyledInput placeholder="new email" />
        </Form.Item>
        <Form.Item<FieldType>
          label="Enter Password"
          name="password"
          rules={[
            { required: true, message: 'password is required' },
          ]}
        >
          <PasswordInput placeholder="password" />
        </Form.Item>
      </StyledForm>
    </Modal>
  );
}

export default ModifyEmailModal;