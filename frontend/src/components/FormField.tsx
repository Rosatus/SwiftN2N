import type {ReactNode} from 'react';

type FormFieldProps = {
    label: string;
    helper?: string;
    children: ReactNode;
}

export function FormField({label, helper, children}: FormFieldProps) {
    return (
        <label className="field">
            <span>{label}</span>
            {children}
            {helper && <small>{helper}</small>}
        </label>
    );
}
