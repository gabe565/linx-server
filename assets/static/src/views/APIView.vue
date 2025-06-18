<script setup>
import UploadHeaders from "@/components/api/UploadHeaders.vue";
import UploadJSONResponse from "@/components/api/UploadJSONResponse.vue";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card/index.js";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { ApiPath } from "@/config/api";
import { AlphaNum, randomString } from "@/util/random.js";
</script>

<template>
  <Card class="container max-w-3xl mx-auto">
    <CardHeader>
      <CardTitle>API Reference</CardTitle>
    </CardHeader>
    <CardContent>
      <Accordion type="single" collapsible class="space-y-4">
        <!-- Upload -->
        <AccordionItem value="upload-choose">
          <AccordionTrigger class="text-lg font-semibold">Upload a File</AccordionTrigger>
          <AccordionContent class="prose space-y-4">
            <p>Send a <code class="font-mono px-1 rounded">PUT</code> request to:</p>
            <pre
              class="overflow-x-auto p-3 rounded text-sm font-mono"
            ><code>{{ ApiPath('/upload') }}/</code></pre>

            <p>
              <strong>Note:</strong> The trailing <code>/</code> is important. This will make curl
              append the filename to the URL. See the examples below.
            </p>

            <UploadHeaders />

            <h4 class="text-lg font-medium">Default Response</h4>
            <p>The URL of the uploaded file.</p>

            <h4 class="text-lg font-medium">JSON Response</h4>
            <UploadJSONResponse full filename="myphoto" />

            <h4 class="text-lg font-medium">Examples</h4>
            <h5 class="font-medium">Basic upload</h5>
            <pre
              class="overflow-x-auto p-3 rounded text-sm font-mono"
            ><code>$ curl {{ ApiPath('/upload/') }} -s -T myphoto.jpg
{{ ApiPath(`/${randomString(8, AlphaNum)}.jpg`) }}</code></pre>

            <h5 class="font-medium">Upload with expiry</h5>
            <pre
              class="overflow-x-auto p-3 rounded text-sm font-mono"
            ><code>$ curl {{ ApiPath('/upload/') }} -s -H 'Linx-Expiry: 20m' -T myphoto.jpg
{{ ApiPath(`/${randomString(8, AlphaNum)}.jpg`) }}</code></pre>

            <h5 class="font-medium">Upload from pipe</h5>
            <pre
              class="overflow-x-auto p-3 rounded text-sm font-mono"
            ><code>$ echo hello | curl {{ ApiPath('/upload/hello.txt') }} -s -T -
{{ ApiPath(`/${randomString(8, AlphaNum)}.txt`) }}</code></pre>
          </AccordionContent>
        </AccordionItem>

        <!-- Overwrite -->
        <AccordionItem value="overwrite">
          <AccordionTrigger class="text-lg font-semibold">Overwrite a File</AccordionTrigger>
          <AccordionContent class="prose space-y-4">
            <p>Upload again with the same filename and include the original deletion key:</p>

            <h4 class="text-lg font-medium">Required Headers</h4>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Header</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow>
                  <TableCell><code>Linx-Delete-Key: mysecret</code></TableCell>
                </TableRow>
              </TableBody>
            </Table>

            <h4 class="text-lg font-medium">Examples</h4>
            <pre
              class="overflow-x-auto p-3 rounded text-sm font-mono"
            ><code>$ curl {{ ApiPath('/upload/myphoto.jpg') }} -s \
    -H 'Linx-Delete-Key: mysecret' -T myphoto.jpg
{{ ApiPath('/myphoto.jpg') }}</code></pre>
          </AccordionContent>
        </AccordionItem>

        <!-- Delete -->
        <AccordionItem value="delete">
          <AccordionTrigger class="text-lg font-semibold">Delete a File</AccordionTrigger>
          <AccordionContent class="prose space-y-4">
            <p>
              Send a <code>DELETE</code> request to the public URL and include the original deletion
              key.
            </p>

            <h4 class="text-lg font-medium">Required Headers</h4>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Header</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow>
                  <TableCell><code>Linx-Delete-Key: mysecret</code></TableCell>
                </TableRow>
              </TableBody>
            </Table>

            <h4 class="text-lg font-medium">Examples</h4>
            <pre
              class="overflow-x-auto p-3 rounded text-sm font-mono"
            ><code>$ curl {{ ApiPath('/myphoto.jpg') }} -X DELETE -H 'Linx-Delete-Key: mysecret'
DELETED</code></pre>
          </AccordionContent>
        </AccordionItem>

        <!-- Retrieve Info -->
        <AccordionItem value="info">
          <AccordionTrigger class="text-lg font-semibold">Retrieve File Info</AccordionTrigger>
          <AccordionContent class="prose space-y-4">
            <p>Send a <code>GET</code> request to the public URL and request a JSON response.</p>

            <h4 class="text-lg font-medium">Required Headers</h4>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Header</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow>
                  <TableCell><code>Accept: application/json</code></TableCell>
                </TableRow>
              </TableBody>
            </Table>
            <h4 class="text-lg font-medium">JSON Response</h4>
            <UploadJSONResponse filename="myphoto" />

            <h4 class="text-lg font-medium">Examples</h4>
            <pre
              class="overflow-x-auto p-3 rounded text-sm font-mono"
            ><code>$ curl {{ ApiPath('/myphoto.jpg') }} -H 'Accept: application/json'</code></pre>
          </AccordionContent>
        </AccordionItem>

        <!-- Client -->
        <AccordionItem value="client">
          <AccordionTrigger class="text-lg font-semibold">Client</AccordionTrigger>
          <AccordionContent class="prose">
            <p>
              For convenience, use
              <a target="_blank" href="https://github.com/andreimarcu/linx-client">linx-client</a>
              to simplify file uploads.
            </p>
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    </CardContent>
  </Card>
</template>
