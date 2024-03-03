import React from "react"
import {DictionaryValue, Value} from "./value.ts"
import Type from "./type.tsx"

interface Props {
    value: DictionaryValue,
    onChange?: (value: Value) => void
}

export default function DictionaryValue({
    value,
    onChange
}: Props) {
    function _onChange(event: React.ChangeEvent<HTMLSelectElement>) {
        const index = Number(event.target.value)
        const key = value.keys[index].value
        if (key.kind === "primitive") {
            onChange?.(key.value)
        }
    }

    return (
        <>
            <Type
                kind={value.kind}
                type={value.type}
                description={value.typeString}
            />
            <select size={2} onChange={_onChange}>
                {value.keys.map((key, index) => (
                    <option key={index} value={index}>{key.description}</option>
                ))}
            </select>
        </>
    )
}
